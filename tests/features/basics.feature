Feature: basic commands testing

    Scenario: get the cli version
    When we run the --version command
    Then we should see "Platform.sh CLI"

    Scenario: run help command
    When we run the help command
    Then we should see "Description: Displays help for a command"

    Scenario Outline: Commands
    When we run the <command> command
        then we should see "<result>"

    Examples: Commands
    | command       | result      |
    | projects      | Title       |
    | auth:info     | username    |
