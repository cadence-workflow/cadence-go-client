# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.3.0] - 2025-07-08
### Added
- Added PollerInitCount and reduced default PollerMaxCount in AutoScalerOptions (#1433)
- Added ExecuteWithVersion and ExecuteWithMinVersion to GetVersion (#1427)
- Added NewBatchFuture and BatchFuture API (#1426)
- Added CronOverlapPolicy to StartWorkflowOptions (#1425)
- Added GetSpanContext and WithSpanContext API (#1423)
- Added ConcurrencyLimit to worker to enable dynamic tuning of concurrencies (#1410)
- Added worker.NewV2 with validation on decision poller count (#1370)
- Added interfaces for worker statistics and debugging capabilities (#1356, #1362, #1363)
- Added FirstRunAt to StartWorkflowOptions (#1360)
- Added PR template for breaking changes (#1351)
- Added methods on Worker to get registered workflows and activities (#1342)
- Added workflow and activity APIs to testsuite (#1343)
- Added option to exclude workflows by type (#1335)
- Added server-like `make build` (#1329)
- Added 85% code coverage requirement (#1325)
- Added wrapper implementations for StartWorkflowExecution and SignalWithStartWorkflowExecution APIs (#1321)
- Added codecov integration and metadata generation (#1320)
- Added documentation for context propagators (#1312)
- Increased test coverage to 85% (#1331, #1316, #1345, #1346, #1349, #1350, #1372, #1373, #1378, #1380, #1382, #1383, #1384, #1385, #1386, #1387, #1388, #1389, #1390, #1391, #1392, #1393, #1395, #1397, #1398, #1401, #1404)

### Changed
- Fixed and enforced linting across the codebase (#1429)
- Changed license to Apache 2.0 (#1422)
- Upgraded internal tooling to support Go 1.24 (no user-facing changes) (#1421)
- Reworked poller auto-scaler logic (#1411)
- Updated cadence-idl submodule (#1408)
- Updated Codecov configuration for the new GitHub organization (#1403)
- Documented a significant caveat for SideEffect functions (#1399)
- Moved licensegen file under internal/tools (#1381)
- Updated the minimum supported Go version to 1.21 (#1379)
- Changed registry API signatures to return info interfaces (#1355)
- Updated compatibility adapter to support new enum values (#1337)
- Bumped golang.org/x/tools and related dependencies to support Go 1.22 (#1336)
- Migrated CI from AWS to Google Kubernetes Engine (#1333)
- Extracted domain client into a separate file (#1348)
- Improved example test by sorting output (#1366)
- Pinned mockery version and regenerated all mocks (#1328)
- Updated client wrappers to include new async APIs (#1327)

### Removed
- Removed worker hardware utilization code (#1400)
- Removed deprecated Fossa integration (#1361)
- Removed strings.Compare usage from example tests (#1367)
- Removed Coveralls integration (#1354)

### Fixed
- Fixed AutoConfigHint population in the mapper (#1415)
- Minor race prevention: do not mutate callers' retry policy (#1413)
- Fixed incorrect nil handling in workflowTaskPoller (#1412)
- Fixed go-generate calling, do more before running tests (#1377)
- Restored race-checking tests (#1376)
- Skipped racy tests (#1375)
- Fixed panics in test activities (#1374)
- Adjusted startedCount assertion in Test_WorkflowLocalActivityWithMockAndListeners (#1353)
- Partially fixed Continue as new case (#1347)
- Fixed unit_test failure detection, and tests for data converters (#1341)
- Fixed coverage metadata commit info (#1323)


## [v1.2.9] - 2024-03-01
### Added
- retract directive for v1.2.8

### Changed
- Revert breaking changes from v1.2.8 (#1315)

## [v1.2.8] - 2024-02-27
- Support two-legged OAuth flow (#1304)
- Expose method to get default worker options (#1311)
- Added CloseTime filter to shadower (#1309)
- Making Workflow and Activity registration optional when they are mocked (#1256)
- Addressing difference in workflow interceptors when using the testkit (#1257)
- remove time.Sleep from tests (#1305)

## [v1.2.7] - 2023-12-6
### Changed
- Upgraded cassandra image to 4.1.3 in docker compose files #1301

### Fixed
- Fixed history size exposure logging #1300

## [v1.2.6] - 2023-11-24
### Added
- Added a new query type `__query_types` #1295
- Added calculate workflow history size and count and expose that to client #1270
- Added honor non-determinism fail workflow policy #1287

## [v1.1.0] - 2023-11-06
### Added
- Added new poller thread pool usage metrics #1275 #1291
- Added metrics tag workflowruntimelength in workflow context #1277
- Added GetWorkflowTaskList and GetActivityTaskList APIs #1292

### Changed
- Updated idl version
- Improved retrieval of binaryChecksum #1279

### Fixed
- Fixed error log #1284
- Fixed in TestEnv workflow interceptor is not propagated correctly for child workflows #1289

## [v1.0.2] - 2023-09-25
### Added
- Add a structured error for non-determinism failures

### Changed
- Do not log when automatic heart beating fails due to cancellations

## [v1.0.1] - 2023-08-14
### Added
- Emit cadence worker's hardware utilization inside worker once per host by @timl3136 in #1260
### Changed
- Updated supported Go version to 1.19
- Log when the automatic heartbeating fails
- Updated golang.org/x/net and github.com/prometheus/client_golang

## [v1.0.0] - 2023-07-12
- add refresh tasks API to client by @mkolodezny in #1162
- Exclude idls subfolder from licencegen tool by @vytautas-karpavicius in #1163
- Upgrade x/sys and quantile to work with Go 1.18 by @Groxx in #1164
- Stop retrying get-workflow-history with an impossibly-short timeout by @Groxx in #1171
- Rewrite an irrational test which changes behavior based on compiler inlining by @Groxx in #1172
- Deduplicate retry tests a bit by @Groxx in #1173
- Prevent local-activity panics from taking down the worker process by @Groxx in #1169
- Moving retryable-err checks to errors.As, moving some to not-retryable by @Groxx in #1167
- Apparently copyright isn't checked by CI by @Groxx in #1175
- Another missed license header by @Groxx in #1176
- Add JitterStart support to client by @ZackLK in #1178
- Simplify worker options configuration value propagation by @shijiesheng in #1179
- Sharing one of my favorite "scopes" in intellij, and making it easier to add more by @Groxx in #1182
- Add poller autoscaler by @shijiesheng in #1184
- add poller autoscaling in activity and decision workers by @shijiesheng in #1186
- Fix bug with workflow shadower: ALL is documented as an allowed Status; test and fix. by @ZackLK in #1187
- upgrade thrift to v0.16.0 and tchannel-go to v1.32.1 by @shijiesheng in #1189
- [poller autoscaler] fix logic to identify empty tasks by @shijiesheng in #1192
- Maintain a stable order of children context, resolves a non-determinism around cancels by @Groxx in #1183
- upgrade fossa cli to latest and remove unused fossa.yml by @shijiesheng in #1196
- Retry service-busy errors after a delay by @Groxx in #1174
- changing dynamic poller scaling strategy. by @mindaugasbarcauskas in #1197
- Fix flaky test by @mindaugasbarcauskas in #1201
- updating go client dependencies. by @mindaugasbarcauskas in #1200
- version metrics by @allenchen2244 in #1199
- Export GetRegisteredWorkflowTypes so I can use in shadowtest. by @ZackLK in #1202
- Add GetUnhandledSignalNames by @longquanzheng in #1203
- Adding go version check when building locally. by @mindaugasbarcauskas in #1209
- update CI go version. by @mindaugasbarcauskas in #1210
- ran "make fmt" by @mindaugasbarcauskas in #1206
- Updating IDL version for go client. by @mindaugasbarcauskas in #1211
- Adding ability to provide cancellation reason to cancelWorkflow API by @mindaugasbarcauskas in #1213
- Expose WithCancelReason and related types publicly, as originally intended by @Groxx in #1214
- Add missing activity logger fields for local activities by @Groxx in #1216
- Modernize makefile like server, split tools into their own module by @Groxx in #1215
- adding serviceBusy tag for transient-poller-failure counter metric. by @mindaugasbarcauskas in #1212
- surface more information in ContinueAsNewError by @shijiesheng in #1218
- Corrected error messages in getValidatedActivityOptions by @jakobht in #1224
- Fix TestActivityWorkerStop: it times out with go 1.20 by @dkrotx in #1223
- Fixed the spelling of replay_test file. by @agautam478 in #1226
- Add more detail to how workflow.Now behaves by @Groxx in #1228
- Part1: Record the data type change scenario for shadower/replayer test suite by @agautam478 in #1227
- Document ErrResultPending's behavioral gap explicitly by @Groxx in #1229
- Added the Activity Registration required failure scenario to replayer test suite by @agautam478 in #1231
- Shift replayer to prefer io.Reader rather than filenames by @Groxx in #1234
- Expose activity registry on workflow replayer by @Groxx in #1232
- Merged the timeout logic for the tests in internal_workers_test.go by @jakobht in #1225
- [error] surface more fields in ContinueAsNew error by @shijiesheng in #1235
- Add and emulate the issues found in the workflows involving coroutines into the replayersuite. by @agautam478 in #1237
- Add the change in branch number case(test) to replayersuite by @agautam478 in #1236
- Locally-dispatched activity test flakiness hopefully resolved by @Groxx in #1240
- Switched to revive, goimports, re-formatted everything by @Groxx in #1233
- Add the case where changing the activities (addition/subtraction/modification in current behavior) in the switch case has no effect on replayer. by @agautam478 in #1238
- Replaced Activity.RegisterWithOptions with replayers own acitivty register by @agautam478 in #1242
- [activity/logging] produce a log when activities time out by @sankari165 in #1243
- Better logging when getting some nondeterministic behaviours by @jakobht in #1245
- make fmt fix by @Groxx in #1246
- Test-suite bugfix: local activity errors were not encoded correctly by @Groxx in #1247
- Extracting the replayer specific utilities into a separate file for readability. by @agautam478 in #1244
- Adding WorkflowType to "Workflow panic" log-message by @dkrotx in #1259
- Adding in additional header to determine a more stable isolation-group by @davidporter-id-au in #1252
- Bump version strings for 1.0 release by @Groxx in #1261

## [v0.19.0] - 2022-01-05
### Added
- Added JWT Authorization Provider. This change includes a dependency that uses v2+ go modules. They no longer match import paths, meaning that we have to **drop support for dep & glide** in order to use this. [#1116](https://github.com/uber-go/cadence-client/pull/1116)
### Changed
- Generated proto type were moved out to [cadence-idl](https://github.com/uber/cadence-idl) repository. This is **BREAKING** if you were using `compatibility` package. In that case you will need to update import path from `go.uber.org/cadence/.gen/proto/api/v1` to `github.com/uber/cadence-idl/go/proto/api/v1` [#1138](https://github.com/uber-go/cadence-client/pull/1138)
### Documentation
- Documentation improvements for `client.SignalWorkflow` [#1151](https://github.com/uber-go/cadence-client/pull/1151)


## [v0.18.5] - 2021-11-09
