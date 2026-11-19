import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import DocumentationCode from '../components/documentation/DocumentationCode'

export interface DocumentationArticle {
  slug: string
  title: string
  group: string
  description: string
  keywords: string
  sections: {
    id: string
    title: string
    content: ReactNode
  }[]
}

const loginExample = `Feature: User login

  Scenario: Successful login
    Given the Email field and Password field are filled
    When the user clicks the Login button
    Then the Dashboard heading is displayed`

export const documentationArticles: DocumentationArticle[] = [
  {
    slug: 'introduction',
    title: 'Introduction',
    group: 'overview',
    description:
      'Learn about Story2Gherkin, its workflow, and the basics of Gherkin scenarios.',
    keywords: 'introduction overview about workflow bdd',
    sections: [
      {
        id: 'tentang',
        title: 'What Is Story2Gherkin?',
        content: (
          <>
            <p>
              Story2Gherkin helps System Analysts turn user
              requirements into Gherkin scenarios, review the
              generated results, and save corrections. The QA
              workspace supports scenario quality evaluation.
            </p>

            <p>
              You can read this documentation without signing
              in. Start with the user guides or learn the
              basics of Gherkin before creating your first
              scenario.
            </p>
          </>
        ),
      },
      {
        id: 'alur-kerja',
        title: 'How Does It Work?',
        content: (
          <ol>
            <li>
              <strong>Enter your requirements.</strong> Fill
              in the User Story, GUI Elements, and Free
              Scenario.
            </li>

            <li>
              <strong>Generate and review.</strong> Check
              the Gherkin output, step structure, and GUI
              element coverage.
            </li>

            <li>
              <strong>Save and download.</strong> Save valid
              corrections, then download the{' '}
              <code>.feature</code> file.
            </li>
          </ol>
        ),
      },
      {
        id: 'mengenal-gherkin',
        title: 'What Is Gherkin?',
        content: (
          <>
            <p>
              Gherkin describes examples of system behavior
              using readable, structured text. A{' '}
              <code>.feature</code> file groups scenarios
              and their steps.
            </p>

            <p>
              To execute these scenarios as automated tests,
              Gherkin steps must be connected to code through{' '}
              <em>step definitions</em> and run using a testing
              tool. Story2Gherkin helps you create and validate
              the scenario text.
            </p>
          </>
        ),
      },
      {
        id: 'contoh-skenario',
        title: 'Example Scenario',
        content: (
          <>
            <p>
              The following example describes a successful
              login. GUI element names are written consistently
              so they are easy to check during review.
            </p>

            <DocumentationCode
              code={loginExample}
              filename="login.feature"
            />
          </>
        ),
      },
      {
        id: 'ketersediaan',
        title: 'Available Features',
        content: (
          <>
            <p>
              System Analysts can use Generate Scenario and
              Review Scenario to create, correct, save, and
              download Gherkin scenarios.
            </p>

            <p>
              QA users can manage reference scenarios through
              Golden Dataset and inspect saved results through
              Evaluation Report.
            </p>

            <aside className="docs-note">
              <strong>Run Evaluation Integration</strong>
              <p>
                Run Evaluation is still being integrated.
                Evaluation Report displays results that have
                already been saved. The report list may be
                empty when no evaluation results are available.
              </p>
            </aside>
          </>
        ),
      },
    ],
  },
  {
    slug: 'getting-started',
    title: 'Getting Started',
    group: 'guides',
    description:
      'Create an account, sign in, and find the workspace for your role.',
    keywords:
      'login sign up register account role system analyst qa start',
    sections: [
      {
        id: 'mulai',
        title: 'Before Signing In',
        content: (
          <p>
            Home, About, and Documentation are available to
            all visitors. The <strong>Get Started</strong>{' '}
            button takes you to the Login page.
          </p>
        ),
      },
      {
        id: 'membuat-akun',
        title: 'Create an Account',
        content: (
          <ol>
            <li>
              Open <Link to="/register">Sign Up</Link>.
            </li>

            <li>
              Enter your name, email address, and password.
            </li>

            <li>
              Choose System Analyst or Quality Assurance
              according to your responsibilities.
            </li>

            <li>
              Complete your registration, then{' '}
              <Link to="/login">log in</Link> with your
              account.
            </li>
          </ol>
        ),
      },
      {
        id: 'memilih-ruang-kerja',
        title: 'Workspaces by Role',
        content: (
          <div className="docs-table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Role</th>
                  <th>Page After Login</th>
                  <th>Purpose</th>
                </tr>
              </thead>

              <tbody>
                <tr>
                  <td>System Analyst</td>
                  <td>Generate</td>
                  <td>Create and review scenarios.</td>
                </tr>

                <tr>
                  <td>Quality Assurance</td>
                  <td>Evaluation</td>
                  <td>
                    Manage golden datasets and inspect
                    saved evaluation reports.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        ),
      },
      {
        id: 'sesi-login',
        title: 'Sessions and Logout',
        content: (
          <>
            <p>
              Select Remember me to retain your session
              according to its validity period. If your session
              expires, sign in again to access your workspace.
            </p>

            <p>
              Logging out returns you to Home. If you have
              unsaved review changes, the application asks for
              confirmation before you leave.
            </p>
          </>
        ),
      },
    ],
  },
  {
    slug: 'generate',
    title: 'Generate Scenario',
    group: 'guides',
    description:
      'Turn a user story and GUI information into draft Gherkin scenarios.',
    keywords:
      'system analyst sa user story gui elements free scenario generate',
    sections: [
      {
        id: 'system-analyst',
        title: 'For System Analysts',
        content: (
          <p>
            Sign in as a System Analyst and open Generate.
            Prepare a clear requirement and the user flow
            you want to test.
          </p>
        ),
      },
      {
        id: 'input-kebutuhan',
        title: 'Enter Scenario Requirements',
        content: (
          <>
            <dl className="docs-definitions">
              <dt>User Story</dt>
              <dd>
                Describe the user, their goal, and why they
                want to achieve it.
              </dd>

              <dt>GUI Elements</dt>
              <dd>
                Enter one element name per field. Use Add
                Element to add another field.
              </dd>

              <dt>Free Scenario</dt>
              <dd>
                Describe the initial conditions, user actions,
                and expected results.
              </dd>
            </dl>

            <aside className="docs-note">
              <strong>Example Login Input</strong>

              <p>
                User Story: As a user, I want to log in so
                that I can access my dashboard.
              </p>

              <p>
                GUI Elements: Email field, Password field,
                Login button, Dashboard heading.
              </p>

              <p>
                Free Scenario: The user fills in the email
                and password, clicks the Login button, then
                sees the Dashboard heading.
              </p>
            </aside>
          </>
        ),
      },
      {
        id: 'menjalankan-generate',
        title: 'Generate Scenarios',
        content: (
          <ol>
            <li>Complete all required fields.</li>

            <li>
              Click the <strong>Generate</strong> button
              and wait for the process to finish.
            </li>

            <li>
              Read the results and check whether they match
              your requirements.
            </li>

            <li>
              Once the results are saved, use the{' '}
              <strong>Review</strong> button to inspect
              them further.
            </li>
          </ol>
        ),
      },
      {
        id: 'hasil-generate',
        title: 'Check the Results',
        content: (
          <>
            <p>
              Review AI-generated results carefully. Check
              the feature name, step order, expected results,
              and consistency of GUI element names.
            </p>

            <DocumentationCode
              code={loginExample}
              filename="login.feature"
            />
          </>
        ),
      },
    ],
  },
  {
    slug: 'review',
    title: 'Review Scenario',
    group: 'guides',
    description:
      'Inspect, correct, save, and download your scenarios.',
    keywords:
      'review edit correct save download feature draft system analyst',
    sections: [
      {
        id: 'memilih-skenario',
        title: 'Select a Scenario',
        content: (
          <p>
            Open Review Scenario from the sidebar, then choose
            a scenario from the dropdown. You can also open
            a specific result using the Review button after
            generation.
          </p>
        ),
      },
      {
        id: 'mengedit',
        title: 'Edit a Scenario',
        content: (
          <p>
            Click <strong>Edit Scenario</strong> to change
            the text. Make sure Feature and Scenario have
            titles, and that every Given, When, and Then
            includes a clear step description.
          </p>
        ),
      },
      {
        id: 'menyimpan',
        title: 'Save Changes',
        content: (
          <>
            <p>
              Click <strong>Save Changes</strong> to run
              validation. Valid corrections become the saved
              version. If the text is invalid, use the reported
              line numbers to correct it. Your draft remains
              available, and the saved version stays unchanged.
            </p>

            <p>
              Coverage below 100% indicates incomplete GUI
              element coverage. You can still save the
              correction if the scenario structure is valid.
            </p>
          </>
        ),
      },
      {
        id: 'mengunduh',
        title: 'Download a .feature File',
        content: (
          <p>
            Download .feature retrieves the latest valid saved
            version. If you are editing a scenario, save your
            corrections first to include them in the download.
          </p>
        ),
      },
      {
        id: 'draft-dan-konflik',
        title: 'Unsaved Changes and Other Tabs',
        content: (
          <>
            <p>
              Leaving the page, switching scenarios, or logging
              out requires confirmation when you have unsaved
              changes.
            </p>

            <p>
              If another tab has updated the scenario, copy
              the draft you want to keep, then reload the
              latest version before combining your corrections.
            </p>
          </>
        ),
      },
    ],
  },
  {
    slug: 'evaluation',
    title: 'Evaluation for QA',
    group: 'guides',
    description:
      'Explore the Quality Assurance workspace and evaluation features.',
    keywords:
      'qa quality assurance golden dataset run evaluation report csv',
    sections: [
      {
        id: 'qa',
        title: 'Quality Assurance Workspace',
        content: (
          <p>
            Signing in as QA takes you to Evaluation.
            This workspace provides access to Golden Dataset,
            Run Evaluation, and Evaluation Report according
            to their availability.
          </p>
        ),
      },
      {
        id: 'status-evaluasi',
        title: 'Feature Availability',
        content: (
          <>
            <p>
              Use Golden Dataset to manage reference scenarios.
              Use Evaluation Report to view saved evaluation
              runs, inspect their details, and download CSV
              reports for completed runs.
            </p>

            <aside className="docs-note">
              <strong>Run Evaluation Integration</strong>
              <p>
                Run Evaluation is still being integrated.
                A completed evaluation run must have saved
                results before its report can be viewed
                and downloaded.
              </p>
            </aside>
          </>
        ),
      },
      {
        id: 'persiapan-qa',
        title: 'Prepare for Evaluation',
        content: (
          <p>
            Learn the Gherkin structure and the Generate and
            Review workflows. Agree on the requirements,
            expected scenarios, and GUI element names with
            your team before evaluating scenario quality.
          </p>
        ),
      },
    ],
  },
  {
    slug: 'gherkin',
    title: 'Gherkin Basics',
    group: 'reference',
    description:
      'Understand Feature, Scenario, Given, When, Then, and And.',
    keywords:
      'gherkin syntax keyword feature scenario given when then and bdd',
    sections: [
      {
        id: 'struktur-file',
        title: '.feature File Structure',
        content: (
          <>
            <p>
              <code>Feature:</code> names a feature and groups
              its scenarios. <code>Scenario:</code> describes
              a concrete example of behavior. Both keywords
              are followed by a colon and a title.
            </p>

            <DocumentationCode
              code={loginExample}
              filename="login.feature"
            />
          </>
        ),
      },
      {
        id: 'keyword-langkah',
        title: 'Step Keywords',
        content: (
          <div className="docs-table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Keyword</th>
                  <th>Meaning</th>
                </tr>
              </thead>

              <tbody>
                <tr>
                  <td><code>Given</code></td>
                  <td>Initial conditions before an action.</td>
                </tr>

                <tr>
                  <td><code>When</code></td>
                  <td>The action or event being tested.</td>
                </tr>

                <tr>
                  <td><code>Then</code></td>
                  <td>The expected observable result.</td>
                </tr>

                <tr>
                  <td><code>And</code></td>
                  <td>Continues the preceding step type.</td>
                </tr>
              </tbody>
            </table>
          </div>
        ),
      },
      {
        id: 'menulis-langkah',
        title: 'Write Clear Steps',
        content: (
          <>
            <p>
              Add a description after each keyword. Use Then
              to describe an observable result. Keep indentation
              consistent, such as two spaces per level.
            </p>

            <p>
              For example,{' '}
              <code>When the user clicks the Login button</code>{' '}
              describes an action, while <code>When</code>{' '}
              alone does not describe a complete step.
            </p>
          </>
        ),
      },
      {
        id: 'dukungan-story2gherkin',
        title: 'Story2Gherkin Support',
        content: (
          <aside className="docs-note">
            <strong>Application Review Rules</strong>

            <p>
              Review uses the English keywords Feature,
              Scenario, Given, When, Then, and And. Each
              Scenario must contain Given, When, and Then
              steps with descriptions in the required order.
            </p>

            <p>
              Background, Rule, But, Scenario Outline, and
              Examples are not currently supported by the
              Story2Gherkin review validator.
            </p>
          </aside>
        ),
      },
    ],
  },
  {
    slug: 'validation',
    title: 'Validation & GUI Coverage',
    group: 'reference',
    description:
      'Understand structure checks, validation errors, and GUI element coverage.',
    keywords:
      'valid invalid warning error structure empty description validation coverage covered',
    sections: [
      {
        id: 'validasi',
        title: 'Gherkin Structure',
        content: (
          <p>
            The panel shows checks for Feature, Scenario,
            Given, When, and Then. Complete parts remain
            highlighted in green even when other steps contain
            errors. VALID means the text passes the application
            review rules; it does not establish that the
            system under test works correctly.
          </p>
        ),
      },
      {
        id: 'kesalahan-langkah',
        title: 'Empty or Invalid Steps',
        content: (
          <>
            <p>
              Error messages identify the line number and
              cause. For example,{' '}
              <strong>Given is missing a step description</strong>{' '}
              asks you to add the initial conditions after
              the Given keyword.
            </p>

            <DocumentationCode
              code={
                'Feature: User login\n\n' +
                '  Scenario: Successful login\n' +
                '    Given\n' +
                '    When\n' +
                '    Then'
              }
              filename="invalid-example.feature"
            />

            <p>
              Complete the three steps, then save again
              to validate your corrections.
            </p>
          </>
        ),
      },
      {
        id: 'gui-coverage',
        title: 'GUI Coverage',
        content: (
          <>
            <p>
              Story2Gherkin calculates the percentage of GUI
              elements found in the step text. Titles,
              comments, and table contents do not count
              as GUI element usage.
            </p>

            <p>
              <code>
                Coverage = used GUI elements / total GUI
                elements × 100%
              </code>
            </p>

            <p>
              If 3 of 4 elements are mentioned, coverage is
              75%. USED lists the elements found, while NOT
              COVERED lists those not found. This percentage
              measures completeness of element references,
              rather than the success of automated tests.
            </p>
          </>
        ),
      },
      {
        id: 'belum-dihitung',
        title: 'Coverage Not Available',
        content: (
          <p>
            If the structure cannot be parsed or a step is
            missing its description, correct the text first.
            The panel indicates that coverage is unavailable
            so this condition is distinguishable from a
            calculated coverage of 0%.
          </p>
        ),
      },
      {
        id: 'coverage-dan-simpan',
        title: 'How Coverage Affects Saving',
        content: (
          <p>
            Coverage below 100% does not prevent saving when
            the structure is valid. Invalid structures are
            rejected. You can correct the draft while the
            latest valid saved version remains available.
          </p>
        ),
      },
    ],
  },
  {
    slug: 'faq',
    title: 'Frequently Asked Questions',
    group: 'help',
    description:
      'Find answers to common questions about using Story2Gherkin.',
    keywords:
      'faq help troubleshooting login download error generate evaluation',
    sections: [
      {
        id: 'akun-dan-akses',
        title: 'Accounts and Access',
        content: (
          <div className="docs-faq">
            <details>
              <summary>
                Do I need to log in to browse the website?
              </summary>

              <p>
                No. Home, About, and Documentation are open
                to all visitors. Login is required to access
                a workspace.
              </p>
            </details>

            <details>
              <summary>
                Why can’t I see Generate or Evaluation?
              </summary>

              <p>
                Navigation follows your account role.
                System Analysts have access to Generate
                and Review, while QA users have access to
                Evaluation. If your session has expired,
                sign in again.
              </p>
            </details>
          </div>
        ),
      },
      {
        id: 'hasil-dan-koreksi',
        title: 'Results and Corrections',
        content: (
          <div className="docs-faq">
            <details>
              <summary>
                Why can’t I save my changes?
              </summary>

              <p>
                Check the validation messages. Complete
                missing step descriptions and correct
                their order. If another tab has updated the
                scenario, keep a copy of your draft before
                reloading.
              </p>
            </details>

            <details>
              <summary>
                Why is my latest edit missing from the download?
              </summary>

              <p>
                Downloads use the saved version. Click
                Save Changes and make sure it succeeds
                before downloading again.
              </p>
            </details>

            <details>
              <summary>
                Does VALID mean the tests have been executed?
              </summary>

              <p>
                No. Review validation checks scenario text.
                Automated testing requires step definitions
                and a testing tool configured for the
                system under test.
              </p>
            </details>
          </div>
        ),
      },
    ],
  },
]

export const documentationGroups = [
  { id: 'overview', label: 'Getting Started' },
  { id: 'guides', label: 'User Guides' },
  { id: 'reference', label: 'Gherkin & Validation' },
  { id: 'help', label: 'Help' },
]

const legacyTopics: Record<string, string> = {
  '#mulai': 'getting-started',
  '#system-analyst': 'generate',
  '#qa': 'evaluation',
  '#validasi': 'validation',
}

export function resolveDocumentationTopic(
  topic: string | null,
  hash: string
) {
  const selected =
    topic || legacyTopics[hash] || 'introduction'

  return (
    documentationArticles.find(
      (article) => article.slug === selected
    ) || documentationArticles[0]
  )
}