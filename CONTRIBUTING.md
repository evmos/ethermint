# Ethermint Contributor Guidelines

* [General Procedure](#general_procedure)
* [Architecture Decision Records (ADR)](#adr)
* [Forking](#forking)
* [Dependencies](#dependencies)
* [Protobuf](#protobuf)
* [Development Procedure](#dev_procedure)
* [Testing](#testing)
* [Updating Documentation](#updating_doc)
* [Branching Model and Release](#braching_model_and_release)
  * [PR Targeting](#pr_targeting)
  * [Pull Requests](#pull_requests)
  * [Process for reviewing PRs](#reviewing_prs)
  * [Pull Merge Procedure](#pull_merge_procedure)
  * [Release Procedure](#release_procedure)

## <span id="general_procedure">General Procedure</span>

Thank you for considering making contributions to Ethermint and related repositories!

Ethermint uses [Tendermint’s coding repo](https://github.com/tendermint/coding) for overall information on repository
workflow and standards.

Contributing to this repo can mean many things such as participating in discussion or proposing code changes. To ensure
a smooth workflow for all contributors, the following general procedure for contributing has been established:

1. Either [open](https://github.com/tharsis/ethermint/issues/new/choose)
   or [find](https://github.com/tharsis/ethermint/issues) an issue you have identified and would like to contribute to
   resolving.
2. Participate in thoughtful discussion on that issue.
3. If you would like to contribute:
    1. If the issue is a proposal, ensure that the proposal has been accepted by the Ethermint team.
    2. Ensure that nobody else has already begun working on the same issue. If someone already has, please make sure to
       contact the individual to collaborate.
    3. If nobody has been assigned the issue and you would like to work on it, make a comment on the issue to inform the
       community of your intentions to begin work. Ideally, wait for confirmation that no one has started it. However,
       if you are eager and do not get a prompt response, feel free to dive on in!
    4. Follow standard Github best practices:
        1. Fork the repo
        2. Branch from the HEAD of `development`(For core developers working within the ethermint repo, to ensure a
           clear ownership of branches, branches must be named with the convention `{moniker}/{issue#}-branch-name`).
        3. Make commits
        4. Submit a PR to `development`
    5. Be sure to submit the PR in `Draft` mode. Submit your PR early, even if it's incomplete as this indicates to the
       community you're working on something and allows them to provide comments early in the development process.
    6. When the code is complete it can be marked `Ready for Review`.
    7. Be sure to include a relevant change log entry in the `Unreleased` section of `CHANGELOG.md` (see file for log
       format).
    8. Please make sure to run `make format` before every commit - the easiest way to do this is having your editor run
       it for you upon saving a file. Additionally, please ensure that your code is lint compliant by running `make lint`
       . There are CI tests built into the Ethermint repository and all PR’s will require that these tests pass before
       they are able to be merged.

**Note**: for very small or blatantly obvious problems (such as typos), it is not required to open an issue to submit a
PR, but be aware that for more complex problems/features, if a PR is opened before an adequate design discussion has
taken place in a github issue, that PR runs a high likelihood of being rejected.

Looking for a good place to start contributing? How about checking out
some [good first issues](https://github.com/tharsis/ethermint/issues?q=label%3A%22good+first+issue%22).

## <span id="adr">Architecture Decision Records (ADR)</span>

When proposing an architecture decision for Ethermint, please create
an [ADR](https://github.com/tharsis/ethermint/blob/main/docs/architecture/README.md) so further discussions can be
made. We are following this process so all involved parties are in agreement before any party begins coding the proposed
implementation. If you would like to see some examples of how these are written refer
to [Tendermint ADRs](https://github.com/tendermint/tendermint/tree/master/docs/architecture).

## <span id="forking">Forking</span>

Please note that Go requires code to live under absolute paths, which complicates forking. While my fork lives
at `https://github.com/tharsis/ethermint`, the code should never exist
at `$GOPATH/src/github.com/tharsis/ethermint`. Instead, we use `git remote` to add the fork as a new remote for the
original repo,`$GOPATH/src/github.com/tharsis/ethermint`, and do all the work there.

For instance, to create a fork and work on a branch of it, you would:

1. Create the fork on github, using the fork button.
2. Go to the original repo checked out locally. (i.e. `$GOPATH/src/github.com/tharsis/ethermint`)
3. `git remote rename origin upstream`
4. `git remote add origin git@github.com:tharsis/ethermint.git`

Now `origin` refers to my fork and `upstream` refers to the ethermint version. So I can `git push -u origin master` to
update my fork, and make pull requests to ethermint from there. Of course, replace `tharsis` with your git handle.

To pull in updates from the origin repo, run:

1. `git fetch upstream`
2. `git rebase upstream/master` (or whatever branch you want)

Please **NO DOT** make Pull Requests from `development`.

## <span id="dependencies">Dependencies</span>

We use [Go 1.15](https://github.com/golang/go/wiki/Modules) Modules to manage dependency versions.

The master branch of every Cosmos repository should just build with `go get`, which means they should be kept up-to-date
with their dependencies, so we can get away with telling people they can just `go get` our software.

Since some dependencies are not under our control, a third party may break our build, in which case we can fall back
on `go mod tidy -v`.

## <span id="protobuf">Protobuf</span>

We use [Protocol Buffers](https://developers.google.com/protocol-buffers) along
with [gogoproto](https://github.com/gogo/protobuf) to generate code for use in Ethermint.

For deterministic behavior around Protobuf tooling, everything is containerized using Docker. Make sure to have Docker
installed on your machine, or head to [Docker's website](https://docs.docker.com/get-docker/) to install it.

For formatting code in `.proto` files, you can run `make proto-format` command.

For linting and checking breaking changes, we use [buf](https://buf.build/). You can use the commands `make proto-lint`
and `make proto-check-breaking` to respectively lint your proto files and check for breaking changes.

To generate the protobuf stubs, you can run `make proto-gen`.

We also added the `make proto-all` command to run all the above commands sequentially.

In order for imports to properly compile in your IDE, you may need to manually set your protobuf path in your IDE's
workspace `settings/config`.

For example, in vscode your `.vscode/settings.json` should look like:

```json
{
  "protoc": {
    "options": [
      "--proto_path=${workspaceRoot}/proto",
      "--proto_path=${workspaceRoot}/third_party/proto"
    ]
  }
}
```

## <span id="dev_procedure">Development Procedure</span>

1. The latest state of development is on `development`.
2. `development` must never
   fail `make lint, make test, make test-race, make test-rpc, make test-solidity, make test-import`
3. No `--force` onto `development` (except when reverting a broken commit, which should seldom happen).
4. Create your feature branch from `development` either on `github.com/tharsis/ethermint`, or your fork (
   using `git remote add origin`).
5. Before submitting a pull request, begin `git rebase` on top of `development`.

## <span id="testing">Testing</span>

Ethermint uses [GitHub Actions](https://github.com/features/actions) for automated testing.

## <span id="updating_doc">Updating Documentation</span>

If you open a PR on the Ethermint repo, it is mandatory to update the relevant documentation in `/docs`. Please refer to
the docs subdirectory and make changes accordingly. Prior to approval, the Code owners/approvers may request some
updates to specific docs.

## <span id="braching_model_and_release">Branching Model and Release</span>

User-facing repos should adhere to the [trunk based development branching model](https://trunkbaseddevelopment.com/).

Libraries need not follow the model strictly, but would be wise to.

Ethermint utilizes [semantic versioning](https://semver.org/).

### <span id="pr_targeting">PR Targeting</span>

Ensure that you base and target your PR on the `development` branch.

All feature additions should be targeted against `development`. Bug fixes for an outstanding release candidate should be
targeted against the release candidate branch.

### <span id="pull_requests">Pull Requests</span>

To accommodate the review process, we suggest that PRs are categorically broken up. Ideally each PR addresses only a
single issue. Additionally, as much as possible code refactoring and cleanup should be submitted as separate PRs from
bug fixes/feature-additions.

### <span id="reviewing_prs">Process for reviewing PRs</span>

All PRs require two Reviews before merge. When reviewing PRs, please use the following review explanations:

1. `LGTM` without an explicit approval means that the changes look good, but you haven't pulled down the code, run tests
   locally and thoroughly reviewed it.
2. `Approval` through the GH UI means that you understand the code, documentation/spec is updated in the right places,
   you have pulled down and tested the code locally. In addition:
    * You must think through whether any added code could be partially combined (DRYed) with existing code.
    * You must think through any potential security issues or incentive-compatibility flaws introduced by the changes.
    * Naming convention must be consistent with the rest of the codebase.
    * Code must live in a reasonable location, considering dependency structures (e.g. not importing testing modules in
      production code, or including example code modules in production code).
    * If you approve of the PR, you are responsible for fixing any of the issues mentioned here.
3. If you are only making "surface level" reviews, submit any notes as `Comments` without adding a review.

### <span id="pull_merge_procedure">Pull Merge Procedure</span>

1. Ensure pull branch is rebased on `development`.
2. Run `make test` to ensure that all tests pass.
3. Squash merge pull request.

### <span id="release_procedure">Release Procedure</span>

1. Start on `development`.
2. Create the release candidate branch `rc/v*` (going forward known as `RC`) and ensure it's protected against pushing
   from anyone except the release manager/coordinator. No PRs targeting this branch should be merged unless exceptional
   circumstances arise.
3. On the `RC` branch, prepare a new version section in the `CHANGELOG.md`. All links must be link-ified:   
   `$ python ./scripts/linkify_changelog.py CHANGELOG.md`  
   Copy the entries into a `RELEASE_CHANGELOG.md`. This is needed so the bot knows which entries to add to the release
   page on github.
4. Kick off a large round of simulation testing (e.g. 400 seeds for 2k blocks).
5. If errors are found during the simulation testing, commit the fixes to `development` and create a new `RC` branch (
   making sure to increment the `rcN`).
6. After simulation has successfully completed, create the release branch (`release/vX.XX.X`) from the `RC` branch.
7. Create a PR to `development` to incorporate the `CHANGELOG.md` updates.
8. Tag the release (use `git tag -a`) and create a release in Github.
9. Delete the `RC` branches.

**Note**: tharsis’s Ethermint team currently cuts releases on a need to have basis. We will announce a more
standardized release schedule as we near production readiness.
