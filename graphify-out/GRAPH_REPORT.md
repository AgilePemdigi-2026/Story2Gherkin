# Graph Report - Story2Gherkin  (2026-11-19)

## Corpus Check
- 193 files · ~146,907 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 42 file(s) not represented in the graph (top: .css 16, (none) 11, .feature 10)

## Summary
- 1535 nodes · 3607 edges · 90 communities (70 shown, 20 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 119 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f9bf192e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- historyFeature
- request_payload
- github.com/gin-gonic/gin.Context
- GenerateResponse
- ScenarioValidation
- AuthService
- context.Context
- validate_gherkin
- package.json
- EvaluationReportDetailResponse
- EvaluationReportPage.tsx
- evalBDDRunRepo
- RegisterDependencies
- GoldenDatasetPage.tsx
- App.tsx
- ReviewScenarioPage.tsx
- GherkinScenario
- goldenFeature
- evalFeature
- reviewFeature
- main.py
- generate_steps.py
- EvaluationRunService
- github.com/google/uuid.UUID
- HistoryDetailPage.tsx
- generator.py
- GenerateRequest
- test_scenarios.py
- evaluation_report_handler_test.go
- PublicHeader.tsx
- compilerOptions
- EvaluationPage.tsx
- auth_handler_test.go
- testing.T
- fixture
- compilerOptions
- JWTService
- Story2Gherkin
- BuildEvaluationReportCSV
- ScenarioGenerator
- test_config.py
- GenerateScenarioPage.tsx
- OpenRouterResponder
- generateFeature
- DocumentationPage.tsx
- GenerationRequest
- registerServiceFeature
- Migrasi AI service ke OpenRouter Nemotron
- Authenticate
- evaluation_run_service_test.go
- github.com/gin-gonic/gin.Engine
- EvaluationRunProgress
- .reset
- Review & Koreksi Skenario
- status
- TestScenarioNumbersAreUniqueForConcurrentCreatesAndSelectedInOwnedList
- github.com/cucumber/godog.ScenarioContext
- Backend BDD: Generation History
- Backend BDD: Golden Dataset
- Instalasi dan Konfigurasi
- .reset
- Backend BDD: Run Evaluation
- when
- goldenBDDRouter
- authFeature
- Documentation
- Backend BDD: Review Scenario
- Story2Gherkin AI service
- TestGoldenDatasetPostgresPersistence
- Navigasi publik dan ruang kerja
- React + TypeScript + Vite
- Pengujian
- Menjalankan Aplikasi
- Run Evaluation Implementation Plan
- tsconfig.json
- bdd/__init__.py
- tests/__init__.py
- .invalidDraft
- github.com/story2gherkin

## God Nodes (most connected - your core abstractions)
1. `historyFeature` - 48 edges
2. `goldenFeature` - 41 edges
3. `reviewFeature` - 38 edges
4. `evalFeature` - 34 edges
5. `GenerateRequest` - 32 edges
6. `User` - 32 edges
7. `validate_gherkin()` - 31 edges
8. `GherkinScenario` - 29 edges
9. `RegisterDependencies()` - 29 edges
10. `ScenarioGenerator` - 27 edges

## Surprising Connections (you probably didn't know these)
- `File` --references--> `GenerateResponse`  [INFERRED]
  docs/superpowers/specs/2026-11-19-run-evaluation-design.md → ai-service/app/schemas/generate.py
- `Task 1: Rule-Based Generator di ai-service` --references--> `require_service_key()`  [INFERRED]
  docs/superpowers/plans/2026-11-19-run-evaluation.md → ai-service/app/api/deps.py
- `Run Evaluation Implementation Plan` --references--> `GenerateRequest`  [INFERRED]
  docs/superpowers/plans/2026-11-19-run-evaluation.md → ai-service/app/schemas/generate.py
- `Task 1: Rule-Based Generator di ai-service` --references--> `GenerateRequest`  [INFERRED]
  docs/superpowers/plans/2026-11-19-run-evaluation.md → ai-service/app/schemas/generate.py
- `Run Evaluation Implementation Plan` --references--> `GenerateResponse`  [INFERRED]
  docs/superpowers/plans/2026-11-19-run-evaluation.md → ai-service/app/schemas/generate.py

## Import Cycles
- None detected.

## Communities (90 total, 20 thin omitted)

### Community 0 - "historyFeature"
Cohesion: 0.05
Nodes (15): RequestStatus, HistoryDetail, HistoryListResponse, HistorySummary, HistoryRepository, NewHistoryRepository(), HistoryService, historySummary() (+7 more)

### Community 1 - "request_payload"
Cohesion: 0.07
Nodes (46): Internal scenario generation service., parametrize, test_application_uses_its_registered_generator(), fake_inference(), test_failure_responses_are_safe_and_release_slot(), test_invalid_input_never_reaches_model(), test_service_authentication_prevents_model_call(), test_startup_loads_settings_when_not_explicitly_injected() (+38 more)

### Community 2 - "github.com/gin-gonic/gin.Context"
Cohesion: 0.10
Nodes (28): authHandler, evaluationReportError(), EvaluationReportHandler, NewEvaluationReportHandler(), evaluationRunError(), EvaluationRunHandler, NewEvaluationRunHandler(), createContext() (+20 more)

### Community 3 - "GenerateResponse"
Cohesion: 0.09
Nodes (34): ScenarioStatus, SourceType, GenerateRequest, GenerateResponse, IsValidRequestStatus(), IsValidScenarioStatus(), IsValidSourceType(), LLMClient (+26 more)

### Community 4 - "ScenarioValidation"
Cohesion: 0.08
Nodes (30): GoldenDatasetDetailResponse, GoldenDatasetListResponse, GoldenDatasetSummary, ReviewScenario, ScenarioSummary, ScenarioValidation, ReviewRepository, NewReviewRepository() (+22 more)

### Community 5 - "AuthService"
Cohesion: 0.07
Nodes (27): NewEmailConfig(), AuthService, authService, newPasswordResetToken(), passwordResetHash(), passwordResetURL(), validNewPassword(), validPasswordResetToken() (+19 more)

### Community 6 - "context.Context"
Cohesion: 0.12
Nodes (15): RefreshToken, Role, User, NewRefreshTokenRepository(), userRepository, NewUserRepository(), bddUserRepo, context.Context (+7 more)

### Community 7 - "validate_gherkin"
Cohesion: 0.09
Nodes (37): post, validate(), BaseModel, ScenarioValidation, StructureChecks, ValidateScenarioRequest, ValidationIssue, inspect_incomplete_source() (+29 more)

### Community 8 - "package.json"
Cohesion: 0.05
Nodes (40): dependencies, lucide-react, react, react-dom, react-router-dom, devDependencies, eslint, @eslint/js (+32 more)

### Community 9 - "EvaluationReportDetailResponse"
Cohesion: 0.10
Nodes (27): EvaluationMethodResult, EvaluationReportDetailResponse, EvaluationReportMetrics, EvaluationRunListResponse, EvaluationRunSummary, evaluationMethodToDTO(), evaluationMetricsToDTO(), evaluationRunToSummary() (+19 more)

### Community 10 - "EvaluationReportPage.tsx"
Cohesion: 0.11
Nodes (36): demoReport, Coverage(), dateFormatter, EvaluationReportPage(), load(), formatDate(), formatPercent(), MetricRow() (+28 more)

### Community 11 - "evalBDDRunRepo"
Cohesion: 0.09
Nodes (14): EvaluationCaseMethodResult, EvaluationRunCase, EvaluationMethodMetrics, EvaluationRun, EvaluationRunMetrics, EvaluationRunStatus, newGoldenBDDStore(), newEvalBDDRunRepo() (+6 more)

### Community 12 - "RegisterDependencies"
Cohesion: 0.11
Nodes (28): SetUpDatabaseConnection(), SetupLogger(), NewScenarioHandler(), NewEvaluationReportRepository(), NewEvaluationRunRepository(), GenerationRequestRepository, NewGenerationRequestRepository(), GherkinScenarioRepository (+20 more)

### Community 13 - "GoldenDatasetPage.tsx"
Cohesion: 0.10
Nodes (30): DatasetFormDialog(), addItem(), submit(), DeleteDialog(), confirm(), DetailDialog(), emptyDraft, formatDate() (+22 more)

### Community 14 - "App.tsx"
Cohesion: 0.13
Nodes (25): App(), router, AuthHeader(), AboutProjectPage(), benefits, objectives, technologies, ForgotPasswordPage() (+17 more)

### Community 15 - "ReviewScenarioPage.tsx"
Cohesion: 0.11
Nodes (24): GherkinEditor(), HighlightedGherkin(), Props, Props, ScenarioSelector(), Props, ValidationPanel(), Feedback (+16 more)

### Community 16 - "GherkinScenario"
Cohesion: 0.11
Nodes (11): GherkinScenario, SaveCorrectionRequest, gherkinScenarioRepository, cloneReviewRow(), reviewBDDRouter(), reviewBDDStore, reviewBDDUsers, reviewBDDValidator (+3 more)

### Community 17 - "goldenFeature"
Cohesion: 0.09
Nodes (3): goldenFeature, bytes.Buffer, encoding/json.RawMessage

### Community 18 - "evalFeature"
Cohesion: 0.10
Nodes (4): errEvalFixture(), InitializeEvaluationScenario(), TestRunEvaluationFeatures(), evalFeature

### Community 20 - "main.py"
Cohesion: 0.15
Nodes (14): get_generator(), Request, require_service_key(), health(), AsyncClient, AsyncClient, Settings, build_client() (+6 more)

### Community 21 - "generate_steps.py"
Cohesion: 0.15
Nodes (22): FakeClient, make_request(), given, then, when, step_dup_lalu_valid(), step_feedback_duplikat(), step_feedback_ghost() (+14 more)

### Community 22 - "EvaluationRunService"
Cohesion: 0.14
Nodes (10): EvaluationRunRepository, GoldenDatasetRepository, NewGoldenDatasetRepository(), caseResult(), EvaluationClient, EvaluationRunService, NewEvaluationRunService(), round2() (+2 more)

### Community 23 - "github.com/google/uuid.UUID"
Cohesion: 0.18
Nodes (8): GoldenDataset, GoldenDatasetItem, lockGoldenDataset(), cloneGoldenDataset(), goldenBDDStore, github.com/google/uuid.UUID, GoldenDatasetItem, fakeDatasetRepo

### Community 24 - "HistoryDetailPage.tsx"
Cohesion: 0.16
Nodes (18): ScenarioSidebar(), formatDate(), formatStatus(), HistoryDetailPage(), loadDetail(), formatDate(), formatStatus(), HistoryPage() (+10 more)

### Community 25 - "generator.py"
Cohesion: 0.27
Nodes (16): generate_rule_based_route(), GeneratedFeature, GenerateResponse, BaseModel, Scenario, StrictModel, render_gherkin(), _chunks() (+8 more)

### Community 26 - "GenerateRequest"
Cohesion: 0.24
Nodes (14): generate(), post, Request, InvalidModelOutput, ModelUnavailable, Respons model ada tapi tidak bisa dipakai: terpotong, bukan Gherkin valid., LLM tidak bisa dihubungi: timeout, rate-limit, downtime provider., OllamaClient (+6 more)

### Community 27 - "test_scenarios.py"
Cohesion: 0.18
Nodes (19): authenticated(), busy(), call_count(), credentials(), forwarded_language(), invalid_request(), language(), limit() (+11 more)

### Community 28 - "evaluation_report_handler_test.go"
Cohesion: 0.29
Nodes (18): assertEvaluationHTTPError(), evaluationHTTPData(), evaluationHTTPPaths(), newEvaluationHTTPFixture(), TestEvaluationReportHTTPDetailPreservesJSONValues(), TestEvaluationReportHTTPDownload(), TestEvaluationReportHTTPHidesStorageErrors(), TestEvaluationReportHTTPInvalidParametersAndMissingRuns() (+10 more)

### Community 29 - "PublicHeader.tsx"
Cohesion: 0.25
Nodes (15): PublicHeader(), RequireRole(), useAuthSession(), HomePage(), UserRole, AuthSession, clearAuthSession(), decodeSession() (+7 more)

### Community 30 - "compilerOptions"
Cohesion: 0.10
Nodes (19): compilerOptions, allowArbitraryExtensions, allowImportingTsExtensions, erasableSyntaxOnly, jsx, lib, module, moduleDetection (+11 more)

### Community 31 - "EvaluationPage.tsx"
Cohesion: 0.18
Nodes (17): Amandemen spec, Task 2: Client Go memanggil generator rule-based, Task 4: Service orkestrasi + metrik, Task 5: DTO, handler, route, DI, Task 6: Halaman Run Evaluation, Task 7: Dokumentasi + smoke test manual, EvaluationPage(), start() (+9 more)

### Community 32 - "auth_handler_test.go"
Cohesion: 0.25
Nodes (15): NewAuthController(), assertLogoutSuccess(), decodeResponse(), doRegister(), newLogoutTestInjector(), newTestHandler(), TestLoginHandler(), TestLogoutHandler() (+7 more)

### Community 33 - "testing.T"
Cohesion: 0.27
Nodes (15): NewAuthService(), TestResetPassword_RejectsInvalidPasswords(), TestResetPassword_RejectsInvalidTokens(), TestResetPassword_RejectsJWT(), TestSendPasswordReset_DBError_IsPreserved(), TestSendPasswordReset_UserNotFound_ReturnsNil(), newTestService(), TestLoginService() (+7 more)

### Community 34 - "fixture"
Cohesion: 0.22
Nodes (15): fixture(), routerFor(), TestDownloadRejectsInvalidStoredVersion(), TestHTTPAuthorization(), TestHTTPDownloadUsesSavedTextAndSafeFilename(), TestHTTPInvalidSaveReturnsLineDetailsWithoutWriting(), TestInvalidCorrectionAndUnavailableValidatorDoNotWrite(), TestRoleAndOwnershipAreEnforcedForEveryReviewOperation() (+7 more)

### Community 35 - "compilerOptions"
Cohesion: 0.12
Nodes (16): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, noEmit, noFallthroughCasesInSwitch (+8 more)

### Community 36 - "JWTService"
Cohesion: 0.16
Nodes (8): getSecretKey(), JWTService, NewJWTService(), time.Duration, jwt.RegisteredClaims, jwt.Token, jwtCustomClaim, loginJWTStub

### Community 37 - "Story2Gherkin"
Cohesion: 0.12
Nodes (15): API, API internal — Python, API publik — Backend Go, Arsitektur, Daftar Isi, Fitur dan Status, Kendala Umum, Kontribusi Tim (+7 more)

### Community 38 - "BuildEvaluationReportCSV"
Cohesion: 0.31
Nodes (13): BuildEvaluationReportCSV(), evaluationCSVMethod(), evaluationCSVPercent(), evaluationCSVText(), evaluationCSVTestPercent(), evaluationCSVTestReport(), readEvaluationCSVTest(), TestEvaluationReportCSVEmptyCases() (+5 more)

### Community 39 - "ScenarioGenerator"
Cohesion: 0.20
Nodes (7): ModelClient, Port ke LLM: mengembalikan teks Gherkin mentah. Parsing + validasi struktur…, extract_feature_block(), Fallback bila model lupa tag @gui-N: cocokkan substring case-insensitive.…, Ambil blok feature dari teks model: mulai dari baris Feature: pertama, buang…, ScenarioGenerator, Protocol

### Community 40 - "test_config.py"
Cohesion: 0.19
Nodes (8): clean_env(), fixture, parametrize, test_invalid_or_paid_model_rejected(), test_invalid_provider_key_rejected(), test_invalid_service_secret_rejected(), test_invalid_timeout_rejected(), test_invalid_token_budget_rejected()

### Community 41 - "GenerateScenarioPage.tsx"
Cohesion: 0.22
Nodes (12): FormErrors, GenerateDraft, GenerateScenarioPage(), getDraftKey(), GuiElement, loadDraft(), renderGherkin(), ApiResponse (+4 more)

### Community 42 - "OpenRouterResponder"
Cohesion: 0.18
Nodes (8): forbid_real_http(), harness_factory(), build(), fixture, No real model/API access is permitted, even if a mock is forgotten., Harness, OpenRouterResponder, Only the external OpenRouter HTTP boundary is replaced in API/BDD tests.

### Community 43 - "generateFeature"
Cohesion: 0.21
Nodes (4): InitializeGenerateScenario(), TestGenerateFeatures(), generateFeature, stubScenarioSvc

### Community 44 - "DocumentationPage.tsx"
Cohesion: 0.23
Nodes (7): DocumentationCode(), DocumentationArticle, documentationArticles, documentationGroups, legacyTopics, resolveDocumentationTopic(), DocumentationPage()

### Community 45 - "GenerationRequest"
Cohesion: 0.24
Nodes (4): GenerationRequest, cloneHistoryRequest(), cloneHistoryScenario(), memReqRepo

### Community 47 - "Migrasi AI service ke OpenRouter Nemotron"
Cohesion: 0.20
Nodes (9): Error provider, Format output, Isi environment lokal, Jalankan, Kontrak Go–Python, Migrasi AI service ke OpenRouter Nemotron, Pengujian, Perubahan kode (+1 more)

### Community 48 - "Authenticate"
Cohesion: 0.33
Nodes (6): Authenticate(), RegisterEvaluationReportRoutes(), RegisterEvaluationRunRoutes(), RegisterGoldenDatasetRoutes(), RegisterRoutes(), github.com/gin-gonic/gin.RouterGroup

### Community 49 - "evaluation_run_service_test.go"
Cohesion: 0.44
Nodes (9): TestCaseFailureDoesNotStopRun(), testCases(), TestExecuteFailsWhenCaseResultCannotBeSaved(), testItems(), TestRunCasesComputesStatusesAndMetrics(), TestStartRejectsNonQA(), TestStartRejectsWhenActiveRunExists(), TestWarnWhenCoverageBelowFull() (+1 more)

### Community 50 - "github.com/gin-gonic/gin.Engine"
Cohesion: 0.25
Nodes (6): main(), run(), CORSMiddleware(), RegisterRoutes(), github.com/gin-gonic/gin.Engine, github.com/gin-gonic/gin.HandlerFunc

### Community 51 - "EvaluationRunProgress"
Cohesion: 0.28
Nodes (4): EvaluationRunCreated, EvaluationRunProgress, EvaluationRunCreateRequest, fakeRunService

### Community 52 - ".reset"
Cohesion: 0.25
Nodes (3): InitializeGoldenDatasetScenario(), TestGoldenDatasetFeatures(), validGoldenFile()

### Community 53 - "Review & Koreksi Skenario"
Cohesion: 0.22
Nodes (8): Alur, Aturan SRS, Checklist browser, Daftar file perubahan, Diagnostik validasi, Kode skenario, Menjalankan dan menguji, Review & Koreksi Skenario

### Community 54 - "status"
Cohesion: 0.29
Nodes (7): status(), Aturan metrik, `GET /api/v1/evaluation-runs/:run_id/progress`, `POST /api/v1/evaluation-runs`, Run Evaluation, API, Task 3: Entity + repository run evaluasi

### Community 55 - "TestScenarioNumbersAreUniqueForConcurrentCreatesAndSelectedInOwnedList"
Cohesion: 0.43
Nodes (5): Migrate(), scenarioCodeDatabase(), TestScenarioCodesAreReturnedAndRemainStableAfterCorrection(), TestScenarioNumberMigrationBackfillsExistingRowsAndDoesNotRenumber(), TestScenarioNumbersAreUniqueForConcurrentCreatesAndSelectedInOwnedList()

### Community 56 - "github.com/cucumber/godog.ScenarioContext"
Cohesion: 0.29
Nodes (5): InitializeRegisterServiceScenario(), TestRegisterServiceFeatures(), InitializeHistoryScenario(), TestHistoryFeatures(), github.com/cucumber/godog.ScenarioContext

### Community 57 - "Backend BDD: Generation History"
Cohesion: 0.29
Nodes (6): Backend BDD: Generation History, Batas Pengujian, Cakupan, Hasil Pengujian, Mekanisme Pengujian, Menjalankan

### Community 58 - "Backend BDD: Golden Dataset"
Cohesion: 0.29
Nodes (6): Backend BDD: Golden Dataset, Batas Pengujian, Cakupan, Hasil Pengujian, Mekanisme Pengujian, Menjalankan

### Community 59 - "Instalasi dan Konfigurasi"
Cohesion: 0.29
Nodes (7): 1. Ambil repository, 2. Siapkan PostgreSQL, 3. Siapkan environment backend, 4. Buat secret dan konfigurasi AI service, 5. Pasang dependency, 6. Siapkan model, Instalasi dan Konfigurasi

### Community 60 - ".reset"
Cohesion: 0.33
Nodes (3): InitializeReviewScenario(), TestReviewFeatures(), reviewValidReport()

### Community 61 - "Backend BDD: Run Evaluation"
Cohesion: 0.33
Nodes (5): Backend BDD: Run Evaluation, Batas Pengujian, Cakupan, Mekanisme, Menjalankan

### Community 62 - "when"
Cohesion: 0.40
Nodes (5): generate(), health(), when, recovery(), unsupported_method()

### Community 63 - "goldenBDDRouter"
Cohesion: 0.60
Nodes (3): goldenBDDRouter(), goldenBDDUsers, goldenBDDValidator

### Community 65 - "Documentation"
Cohesion: 0.40
Nodes (4): Cakupan isi, Documentation, Navigasi, Pemeriksaan

### Community 66 - "Backend BDD: Review Scenario"
Cohesion: 0.40
Nodes (4): Backend BDD: Review Scenario, Batas pengujian, Cakupan, Menjalankan

### Community 67 - "Story2Gherkin AI service"
Cohesion: 0.50
Nodes (3): Jalankan dengan Ollama lokal, Kontrak input generator, Story2Gherkin AI service

### Community 68 - "TestGoldenDatasetPostgresPersistence"
Cohesion: 0.83
Nodes (3): goldenTestDatabase(), goldenTestItems(), TestGoldenDatasetPostgresPersistence()

### Community 69 - "Navigasi publik dan ruang kerja"
Cohesion: 0.50
Nodes (3): Navigasi publik dan ruang kerja, Sesi dan akses, Verifikasi

### Community 70 - "React + TypeScript + Vite"
Cohesion: 0.50
Nodes (3): Expanding the ESLint configuration, React Compiler, React + TypeScript + Vite

### Community 71 - "Pengujian"
Cohesion: 0.50
Nodes (4): Frontend, Go — Unit test dan BDD, Pengujian, Python — Setelah suite pengujian ditambahkan

### Community 72 - "Menjalankan Aplikasi"
Cohesion: 0.50
Nodes (4): Menjalankan Aplikasi, Terminal 1 — AI service, Terminal 2 — Backend, Terminal 3 — Frontend

## Knowledge Gaps
- **185 isolated node(s):** `github.com/story2gherkin`, `EvaluationRunCreateRequest`, `LLMClient`, `UserUpdateRequest`, `UserUpdateResponse` (+180 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 376 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Amandemen spec` connect `EvaluationPage.tsx` to `Run Evaluation Implementation Plan`, `generator.py`, `status`?**
  _High betweenness centrality (0.092) - this node is a cross-community bridge._
- **Why does `Task 1: Rule-Based Generator di ai-service` connect `generator.py` to `GenerateRequest`, `main.py`, `EvaluationPage.tsx`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **What connects `github.com/story2gherkin`, `EvaluationRunCreateRequest`, `LLMClient` to the rest of the system?**
  _185 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `historyFeature` be split into smaller, more focused modules?**
  _Cohesion score 0.051577152600170505 - nodes in this community are weakly interconnected._
- **Should `request_payload` be split into smaller, more focused modules?**
  _Cohesion score 0.06573426573426573 - nodes in this community are weakly interconnected._
- **Should `github.com/gin-gonic/gin.Context` be split into smaller, more focused modules?**
  _Cohesion score 0.09994155464640561 - nodes in this community are weakly interconnected._
- **Should `GenerateResponse` be split into smaller, more focused modules?**
  _Cohesion score 0.08974358974358974 - nodes in this community are weakly interconnected._