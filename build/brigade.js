const { events, Job, Group } = require("brigadier");
const { Check } = require("@brigadecore/brigade-utils");

// **********************************************
// Globals
// **********************************************

const projectName = "porter-terraform";

// **********************************************
// Event Handlers
// **********************************************

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", runSuite);
events.on("issue_comment:created", (e, p) => Check.handleIssueComment(e, p, runSuite));
events.on("issue_comment:edited", (e, p) => Check.handleIssueComment(e, p, runSuite));

events.on("exec", (e, p) => {
  Group.runAll([
    build(e, p),
    xbuild(e, p),
    test(e, p),
  ]);
});

// Although a GH App will trigger 'check_suite:requested' on a push to main event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  if (e.revision.ref.includes("refs/heads/main") || e.revision.ref.startsWith("refs/tags/")) {
    publish(e, p).run();
  }
});

events.on("publish", (e, p) => {
  publish(e, p).run();
});

// **********************************************
// Actions
// **********************************************

function build(e, p) {
  var goBuild = new GoJob(`${projectName}-build`);

  goBuild.tasks.push(
    "make build"
  );

  return goBuild;
}

function xbuild(e, p) {
  var goBuild = new GoJob(`${projectName}-xbuild`);

  goBuild.tasks.push(
    "make xbuild-all"
  );

  return goBuild;
}

function test(e, p) {
  var goTest = new GoJob(`${projectName}-test`);

  goTest.tasks.push(
    "make test-unit"
  );

  return goTest;
}

function publish(e, p) {
  var goPublish = new GoJob(`${projectName}-publish`);

  // TODO: we could/should refactor so that this job shares a mount with the xbuild job above,
  // to remove the need of re-xbuilding before publishing

  goPublish.env.AZURE_STORAGE_CONNECTION_STRING = p.secrets.azureStorageConnectionString;
  goPublish.tasks.push(
    "make xbuild-all publish"
  )

  return goPublish;
}

// Here we add GitHub Check Runs, which will run in parallel and report their results independently to GitHub
function runSuite(e, p) {
  // Important: To prevent Promise.all() from failing fast, we catch and
  // return all errors. This ensures Promise.all() always resolves. We then
  // iterate over all resolved values looking for errors. If we find one, we
  // throw it so the whole build will fail.
  //
  // Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all#Promise.all_fail-fast_behaviour
  //
  // Note: as provided language string is used in job naming, it must consist
  // of lowercase letters and hyphens only (per Brigade/K8s restrictions)
  return Promise.all([
    runTests(e, p, build).catch((err) => { return err }),
    runTests(e, p, xbuild).catch((err) => { return err }),
    runTests(e, p, test).catch((err) => { return err }),
  ])
    .then((values) => {
      values.forEach((value) => {
        if (value instanceof Error) throw value;
      });
    });
}

// **********************************************
// Classes/Helpers
// **********************************************

// GoJob is a Job with Golang-related prerequisites set up
class GoJob extends Job {
  constructor (name) {
    super(name);

    const gopath = "/go";
    const localPath = gopath + `/src/get.porter.sh/mixin/${projectName}`;

    // Here using the large-but-useful deis/go-dev image as we have a need for deps
    // already pre-installed in this image, e.g. helm, az, docker, etc.
    // TODO: replace with lighter-weight image (Carolyn)
    this.image = "deis/go-dev";
    this.env = {
      "GOPATH": gopath
    };
    this.tasks = [
      // Need to move the source into GOPATH so vendor/ works as desired.
      "mkdir -p " + localPath,
      "mv /src/* " + localPath,
      "mv /src/.git " + localPath,
      "cd " + localPath
    ];
    this.streamLogs = true;
  }
}

// runCheck is the default function invoked on a check_run:* event
//
// It determines which check is being requested (from the payload body)
// and runs this particular check, or else throws an error if the check
// is not found
function runCheck(e, p) {
  payload = JSON.parse(e.payload);

  // Extract the check name
  name = payload.body.check_run.name;

  // Determine which check to run
  switch (name) {
    case `${projectName}-build`:
      return runTests(e, p, build);
    case `${projectName}-xbuild`:
      return runTests(e, p, xbuild);
    case `${projectName}-test`:
      return runTests(e, p, test);
    default:
      throw new Error(`No check found with name: ${name}`);
  }
}

// runTests is a Check Run that is run as part of a Checks Suite
function runTests(e, p, jobFunc) {
  console.log("Check requested");

  var check = new Check(e, p, jobFunc(),
    `https://brigadecore.github.io/kashti/builds/${e.buildID}`);
  return check.run();
}

