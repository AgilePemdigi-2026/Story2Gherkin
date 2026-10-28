@bdd
Feature: Generate draft Gherkin scenarios from requirements
  As a backend caller
  I want the AI service to validate generated scenarios
  So that I receive a structured and parseable draft for human review

  Background:
    Given the AI service has a valid model response
    And the caller has valid service credentials
    And a valid user story and two acceptance criteria are provided

  Scenario: Generate positive and negative scenarios
    When the caller requests scenario generation
    Then the HTTP status is 200
    And the Gherkin output contains 2 parseable scenarios
    And both positive and negative scenarios are returned
    And model metadata and a prompt version are returned
    And the model is called 1 times
    And the response does not expose internal secrets

  Scenario: Warn when a criterion is not covered
    Given the model behavior is "uncovered"
    When the caller requests scenario generation
    Then the HTTP status is 200
    And the response warns that criterion 2 is uncovered

  Scenario: Forward the requested content language
    Given the requested language is "id"
    When the caller requests scenario generation
    Then the HTTP status is 200
    And the model receives the requested language "id"

  Scenario Outline: Reject invalid requirements before model inference
    Given the request has the invalid field "<field>"
    When the caller requests scenario generation
    Then the HTTP status is 422
    And no model call is made

    Examples:
      | field             |
      | short_story       |
      | empty_criteria    |
      | excessive_limit   |
      | unknown_language  |
      | unexpected_field  |

  Scenario Outline: Reject unauthorized callers
    Given the caller credentials are "<credentials>"
    When the caller requests scenario generation
    Then the HTTP status is 401
    And no model call is made

    Examples:
      | credentials |
      | missing     |
      | incorrect   |

  Scenario: Reject an output exceeding the requested scenario limit
    Given the requested scenario limit is 1
    When the caller requests scenario generation
    Then the HTTP status is 502
    And the model is called 1 times
    And the response does not expose internal secrets
