import type { EvaluationReport } from '../services/evaluationReportApi'

export const demoReport: EvaluationReport = {
  run_id: 'demo',
  report_code: 'ER-0000',
  dataset_name: 'Golden Dataset v1',
  created_at: '2026-11-17T09:00:00+07:00',
  total_cases: 4,
  status: 'COMPLETED',
  error_message: null,
  metrics: {
    llm: {
      validation_percent: 95,
      gui_coverage_percent: 91,
      acceptance_criteria_precision_percent: 92,
      acceptance_criteria_recall_percent: 88,
      acceptance_criteria_f1_percent: 90,
    },
    rule_based: {
      validation_percent: 80,
      gui_coverage_percent: 72,
      acceptance_criteria_precision_percent: 80,
      acceptance_criteria_recall_percent: 85,
      acceptance_criteria_f1_percent: 82.42,
    },
  },
  cases: [
    {
      case_id: 'GC-001',
      title: 'Successful login',
      expected_gherkin: 'Feature: Login\n  Scenario: Valid login\n    Given the user opens the login page\n    When the user submits valid credentials\n    Then the dashboard is displayed',
      llm: { status: 'PASS', gui_coverage_percent: 100, acceptance_criteria_precision_percent: 100, acceptance_criteria_recall_percent: 100, acceptance_criteria_f1_percent: 100 },
      rule_based: { status: 'PASS', gui_coverage_percent: 98, acceptance_criteria_precision_percent: 90, acceptance_criteria_recall_percent: 100, acceptance_criteria_f1_percent: 94.74 },
    },
    {
      case_id: 'GC-002',
      title: 'User registration',
      expected_gherkin: 'Feature: Register\n  Scenario: Valid registration\n    Given the visitor opens the registration form\n    When the visitor submits valid account details\n    Then the account is created',
      llm: { status: 'PASS', gui_coverage_percent: 90, acceptance_criteria_precision_percent: 100, acceptance_criteria_recall_percent: 80, acceptance_criteria_f1_percent: 88.89 },
      rule_based: { status: 'FAIL', gui_coverage_percent: 60, acceptance_criteria_precision_percent: 50, acceptance_criteria_recall_percent: 40, acceptance_criteria_f1_percent: 44.44 },
    },
    {
      case_id: 'GC-003',
      title: 'Product checkout',
      expected_gherkin: 'Feature: Checkout\n  Scenario: Place an order\n    Given the cart contains a product\n    When the buyer confirms the order\n    Then the order confirmation is displayed',
      llm: { status: 'WARN', gui_coverage_percent: 74, acceptance_criteria_precision_percent: 75, acceptance_criteria_recall_percent: 75, acceptance_criteria_f1_percent: 75 },
      rule_based: { status: 'FAIL', gui_coverage_percent: 30, acceptance_criteria_precision_percent: 33.33, acceptance_criteria_recall_percent: 25, acceptance_criteria_f1_percent: 28.57 },
    },
    {
      case_id: 'GC-004',
      title: 'Delete user',
      expected_gherkin: 'Feature: Manage users\n  Scenario: Delete an existing user\n    Given the admin opens the user list\n    When the admin confirms deletion\n    Then the user is removed from the list',
      llm: { status: 'PASS', gui_coverage_percent: 100, acceptance_criteria_precision_percent: 92, acceptance_criteria_recall_percent: 100, acceptance_criteria_f1_percent: 95.83 },
      rule_based: { status: 'PASS', gui_coverage_percent: 100, acceptance_criteria_precision_percent: 80, acceptance_criteria_recall_percent: 80, acceptance_criteria_f1_percent: 80 },
    },
  ],
}
