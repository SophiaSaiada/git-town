Feature: observing a parked branch

  Background:
    Given the current branch is a parked branch "branch"
    And an uncommitted file
    When I run "git-town observe"

  Scenario: result
    Then it runs no commands
    And the current branch is still "branch"
    And branch "branch" is now observed
    And there are now no parked branches
    And the uncommitted file still exists

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH | COMMAND       |
      | branch | git add -A    |
      |        | git stash     |
      |        | git stash pop |
    And the current branch is still "branch"
    And branch "branch" is now parked
    And there are now no observed branches
    And the uncommitted file still exists
