Feature: basic commands testing

    Scenario: get the cli version
    When we run the --version command
    Then we should see "Platform.sh CLI"

    @legacycli
    Scenario: run help command
    When we run the help command
    Then we should see "/usr/local/bin/platform"

    @gocli
    Scenario: run help command
    When we run the help command
    Then we should see "/tmp/psh-go"

    Scenario Outline: Commands
    When we run the <command> command
        then we should see "<result>"

    Examples: Commands
    | command       | result      |
    | projects      | Title       |
    | auth:info     | username    |
    