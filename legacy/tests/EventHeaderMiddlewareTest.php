<?php

declare(strict_types=1);

namespace Platformsh\Cli\Tests;

use GuzzleHttp\Psr7\Request;
use GuzzleHttp\Psr7\Response;
use PHPUnit\Framework\TestCase;
use Platformsh\Cli\EventHeaderMiddleware;
use Platformsh\Cli\Service\Config;

class EventHeaderMiddlewareTest extends TestCase
{
    private string $configFile;

    protected function setUp(): void
    {
        $this->configFile = __DIR__ . '/data/mock-cli-config.yaml';
    }

    public function testMiddlewareAddsEventHeader(): void
    {
        putenv('MOCK_CLI_EVENT_NAME=backup:restore');
        try {
            $config = new Config([], $this->configFile);
            $middleware = new EventHeaderMiddleware($config);

            $request = new Request('GET', 'https://api.example.com/test');

            // Create a mock handler that captures the request.
            $capturedRequest = null;
            $handler = function (Request $req, array $options) use (&$capturedRequest) {
                $capturedRequest = $req;
                return new Response(200);
            };

            $wrappedHandler = $middleware($handler);
            $wrappedHandler($request, []);

            $this->assertNotNull($capturedRequest);
            $this->assertEquals('backup:restore', $capturedRequest->getHeaderLine('X-CLI-Event'));
        } finally {
            putenv('MOCK_CLI_EVENT_NAME');
        }
    }

    public function testMiddlewareDoesNotAddHeaderWhenEventNameIsEmpty(): void
    {
        // Ensure no event name is set.
        putenv('MOCK_CLI_EVENT_NAME');

        $config = new Config([], $this->configFile);
        $middleware = new EventHeaderMiddleware($config);

        $request = new Request('GET', 'https://api.example.com/test');

        $capturedRequest = null;
        $handler = function (Request $req, array $options) use (&$capturedRequest) {
            $capturedRequest = $req;
            return new Response(200);
        };

        $wrappedHandler = $middleware($handler);
        $wrappedHandler($request, []);

        $this->assertNotNull($capturedRequest);
        $this->assertFalse($capturedRequest->hasHeader('X-CLI-Event'));
    }

    public function testMiddlewarePreservesExistingHeaders(): void
    {
        putenv('MOCK_CLI_EVENT_NAME=project:info');
        try {
            $config = new Config([], $this->configFile);
            $middleware = new EventHeaderMiddleware($config);

            $request = new Request('GET', 'https://api.example.com/test', [
                'Authorization' => 'Bearer token123',
                'Content-Type' => 'application/json',
            ]);

            $capturedRequest = null;
            $handler = function (Request $req, array $options) use (&$capturedRequest) {
                $capturedRequest = $req;
                return new Response(200);
            };

            $wrappedHandler = $middleware($handler);
            $wrappedHandler($request, []);

            $this->assertNotNull($capturedRequest);
            $this->assertEquals('project:info', $capturedRequest->getHeaderLine('X-CLI-Event'));
            $this->assertEquals('Bearer token123', $capturedRequest->getHeaderLine('Authorization'));
            $this->assertEquals('application/json', $capturedRequest->getHeaderLine('Content-Type'));
        } finally {
            putenv('MOCK_CLI_EVENT_NAME');
        }
    }
}
