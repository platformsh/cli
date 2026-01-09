<?php

declare(strict_types=1);

namespace Platformsh\Cli;

use Platformsh\Cli\Service\Config;
use Psr\Http\Message\RequestInterface;

/**
 * Guzzle middleware that adds an X-CLI-Event header to requests.
 *
 * The event name is typically the command name (e.g., "backup:restore") and
 * is used for analytics tracking via Pendo.
 */
class EventHeaderMiddleware
{
    public function __construct(private readonly Config $config) {}

    public function __invoke(callable $next): callable
    {
        return function (RequestInterface $request, array $options) use ($next) {
            $eventName = $this->config->getEventName();
            if ($eventName !== null) {
                $request = $request->withHeader('X-CLI-Event', $eventName);
            }
            return $next($request, $options);
        };
    }
}
