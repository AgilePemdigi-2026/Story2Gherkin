// Fill these details using verified operational settings before publishing.
export const policyDetails = {
  operator: '[Operator name]',
  contact: '[Contact email]',
  effectiveDate: '[Effective date]',
  providers: '[AI and hosting providers; processing locations; AI input retention and training use]',
  retention: '[Account/content/backup retention periods and deletion request procedure]',
  access: '[Which data SA, QA and administrators can access, including shared datasets/reports]',
  technicalData: '[Technical logs, authentication storage, analytics and verified security measures]',
}

export const legalDocument = {
  title: 'Terms of Use & Privacy Policy',
  sections: [
    {
      title: 'Terms of Use',
      items: [
        `Story2Gherkin, operated by ${policyDetails.operator}, helps generate, review and evaluate Gherkin scenarios. By using the service, you agree to these Terms.`,
        'Provide accurate account information, keep your credentials confidential and use only accounts, roles and content you are authorized to access.',
        'Use the service lawfully. Do not submit authentication secrets, upload malicious files, bypass security controls or access another user’s data without permission.',
        'You retain your rights in submitted content and permit its processing to provide the service. Review generated scenarios and evaluation results: automated output may be incomplete or incorrect and does not guarantee test coverage or software security.',
        `Availability is not guaranteed; keep copies of important results. Access may be restricted for misuse. Contact ${policyDetails.contact} for questions or account closure. These Terms do not exclude rights that cannot lawfully be excluded.`,
      ],
    },
    {
      title: 'Privacy Policy',
      items: [
        'The service processes account details, submitted stories and datasets, generated scenarios, reviews and evaluation history to provide its features, manage access, resolve errors and protect the service.',
        `Content may be processed by service providers. Provider and AI processing details: ${policyDetails.providers}. Submit personal or confidential information only when necessary and appropriately authorized.`,
        `Visibility and access: ${policyDetails.access}. Technical data, browser storage and security: ${policyDetails.technicalData}. No system can guarantee absolute security.`,
        `Retention and deletion: ${policyDetails.retention}. Information may also be retained or disclosed where required by applicable law.`,
        `Subject to applicable law, request access, correction, deletion or withdrawal of consent by contacting ${policyDetails.contact}. The effective date identifies the current version; material changes should be communicated before they apply.`,
      ],
    },
  ],
}
