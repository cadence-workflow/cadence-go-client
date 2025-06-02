# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Added worker.NewV2 with validation on decision poller count (#1370)
- Upgraded internal tooling to support Go 1.24, should be no user-noticeable changes (#1421)

### Fixed
- **Significant** test-suite-only change: `RetryPolicy` handling in tests has been off by 1 for many years.
  This is now fixed, and your tests will likely need to be updated. (#1414)

## [v1.2.10] - 2024-07-10
### Added
- Revert "Handle panics while polling for tasks (#1352)" (#1357)
- Remove coveralls integration (#1354)
- Change registry Apis signature to return info interface (#1355)
- Adjust startedCount assertion in Test_WorkflowLocalActivityWithMockAndListeners (#1353)
- Handle panics while polling for tasks (#1352)
- Ensure PR description follows a template when potential breaking changes are made (#1351)
- Add tests for replayer utils isDecisionMatchEvent (#1350)
- Adding tests for internal_workflow_client (#1349)
- Extracting domain client to a separate file (#1348)
- Test for GetWorkflowHistory (#1346)
- Added test for TerminateWorkflow in the internal package (#1345)
- Implement the registered workflows and activities APIs in testsuite (#1343)
- Add methods on Worker to get registered workflows and activities (#1342)
- Update compability adapter to support new enum value (#1337)
- Bump x/tools for tools, to support go 1.22 (#1336)
- Added an option to exclude the list of workflows by Type (#1335)
- Migrate CI from AWS queues to Google Kubernetes Engine queues (#1333)
- Internal workflow client test improvements (#1331)
- Update client wrappers with new async APIs (#1327)
- Server-like `make build` and ensuring builds are clean in CI (#1329)
- Pin mockery and regenerate everything (#1328)
- Enforce 85% new line coverage (#1325)
- Add documentation for propagators and how they are executed (#1312)
- Update idl and add wrapper implementaton for async start/signalwithstart APIs (#1321)
- Enable codecov and generate metadata file as artifact (#1320)
- Release v1.2.9 (#1317)

### Fixed
- Partial fix for Continue as new case (#1347)
- Fixing unit_test failure detection, and tests for data converters (#1341)
- Fix coverage metadata commit info (#1323)


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
