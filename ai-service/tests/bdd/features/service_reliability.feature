@bdd
Feature: AI service failures and recovery
  As a backend caller
  I want safe and predictable error responses
  So that I can report failures without exposing model details

  Background:
    Given the AI service has a valid model response
    And the caller has valid service credentials
    And a valid user story and two acceptance criteria are provided

  Scenario Outline: Reject invalid model output
    Given the model behavior is "<behavior>"
    When the caller requests scenario generation
    Then the HTTP status is 502
    And the model is called 1 times
    And the response does not expose internal secrets

    Examples:
      | behavior            |
      | invalid_model_json  |
      | invalid_schema      |
      | duplicate           |
      | unknown_criterion   |
      | multiline           |
      | oversized           |

  Scenario Outline: Translate model failures into HTTP errors
    Given the model behavior is "<behavior>"
    When the caller requests scenario generation
    Then the HTTP status is <status>
    And the model is called 1 times
    And the response does not expose internal secrets

    Examples:
      | behavior         | status |
      | unavailable      | 503    |
      | connection_error | 503    |
      | timeout          | 504    |
      | internal_error   | 500    |
      | rate_limited     | 503    |
      | unauthorized     | 502    |
      | forbidden        | 502    |
      | invalid_parameters | 502  |

  Scenario: Reject a request while the generator is busy
    Given the generator slot is occupied
    When the caller requests scenario generation
    Then the HTTP status is 503
    And no model call is made

  Scenario: Release the generator slot after an upstream failure
    Given the model behavior is "unavailable"
    When the caller requests scenario generation
    Then the HTTP status is 503
    When the model recovers
    And the caller requests scenario generation
    Then the HTTP status is 200
    And the model is called 2 times

  Scenario: Health remains a liveness check when the model is unavailable
    Given the model behavior is "unavailable"
    When the caller checks service health
    Then the HTTP status is 200
    And no model call is made

  Scenario: Reject an unsupported generation method
    When the caller requests generation using GET
    Then the HTTP status is 405
    And no model call is made
