Migrate to App-SRE Build Pipeline and Clowder
=============================================

Deployment and configuration of an app on cloud.redhat.com becomes much simpler
after migrating to Clowder because a lot of operational decisions are made for
the app, e.g. logging and kafka topic configuration. The migration involves some
work, of course:  apps must ensure conformity to the conventions enforced by
Clowder before they can be managed by it.

This migration combines two migrations into one: 

* Migrate build pipelines to app-interface
* Migrate apps to Clowder

Performing both migrations together reduces overall work, though you need to
perform more steps before seeing results.

Ensure code repo has a Dockerfile
---------------------------------

App SRE's build conventions require that all images be built using a Dockerfile.  
The Dockerfile can live anywhere in your code repo; you can configure a custom
location in your build_deploy.sh (described later) if you place it somewhere
besides the root folder.

Note that a Dockerfile must not pull from Dockerhub.  App SRE blocks all
requests to Dockerhub due to strict rate limiting imposed on their APIs.

Code changes to consume configuration
-------------------------------------

One of Clowder's key features is centralized configuration.  Instead of cobbling
together an app's configuration from a disparate set of secrets, environment
variables, and ConfigMaps that potentially change from environment to
environment, Clowder combines much of an app's configuration into a single JSON
document and mounts it in the app's container.  This should also insulate apps
from differences between environments, e.g. production, ephemeral, and local
development.

There is a companion client library for Clowder, currently implemented in Go and
Python, that consumes the configuration document mounted into every application
container and exposes it via an API.  This API is the recommended way to consume
configuration that comes from Clowder.

Until a dev team is confident an app will not need to be deployed without
Clowder, please use an environment variable to switch between consuming
configuration from Clowder and from its current configuration method (e.g. env
vars, ConfigMap).

Here are the items that you should consume from the Clowder client library:

* Dependent service hostnames: Look these up by the app name
* Kafka bootstrap URL: Multiple URLs can be provided, though only one is ever
  present today
* Kafka topic names: Please look up the actual topic name based on the requested
  name.
* Web prefix and port number
* Metrics path and port number

There are a couple of less trivial changes that may need to be made, depending
on what services are consumed by an app.

If object storage, i.e. S3, is used by an app, it is recommended that an app
switch to the MinIO client library.  MinIO is used in pre-production
environments, and it also supports interacting with S3.  Thus switching to this
library will allow an app to have to include only one object storage client
library.

Clowder can provision Redis on behalf of an app.  If an app uses Redis, we
suggest testing with the version of Redis deployed by Clowder to ensure it is
compatible.  If not, changes to the app will need to be made.

Develop ClowdApp resource for target service
--------------------------------------------

* Write migration script
* All deployments from one code repo should map to one ClowdApp
* Pod spec can be extracted from existing deployment
* Additional information needed:

    * List of kafka topics
    * Optionally request a PostgreSQL database
    * List of object store buckets
    * Optionally request an in-memory database (i.e. Redis)
    * List other app dependencies (e.g. RBAC)

Add build_deploy.sh and pr_check.sh to source code repo
-------------------------------------------------------

App SRE's build jobs largely rely on shell scripts in the target code repo to
execute the build and tests, respectively.  There are two jobs for each app:
"build master" and "PR check", and each job has a corresponding shell script:
build_deploy.sh and pr_check.sh.

build_deploy.sh builds an app's image using a Dockerfile and pushes to quay with
credentials provided in Jenkins job environment.  Make sure to push the latest
and qa image tags if e2e-deploy backwards compatibility is needed.  There is
little variation in this file between projects, thus there are many examples to
pull from.

pr_check.sh is where an app's unit test, static code analysis, linting, and
smoke/integration testing will be performed.  It is largely up to app owners
what goes into this script.  Smoke/integration testing will be performed by
bonfire, and there is an example script to paste into your app's script.  There
are a few environment variables to plug in at the top for an app, and the rest
of the script should be left untouched.

Both files live in the root folder of source code repo, unless overridden in the
Jenkins job definition (see below).

Create "PR check" and "build master" jenkins jobs in app-interface
------------------------------------------------------------------

Two Jenkins jobs need to be defined for each app in app-interface: one to build
the image and one to run test validations against PRs.

App SRE uses Jenkins Job Builder (JJB) to define jobs in YAML.  Jobs are created
by referencing job templates and filling in template parameters.  There are two
common patterns: one for github repos and another for gitlab repos.

Github:

.. code-block:: yaml

    project:
      name: puptoo-stage
      label: insights
      node: insights
      gh_org: RedHatInsights
      gh_repo: insights-puptoo
      quay_org: cloudservices
      jobs:
      - "insights-gh-pr-check":
          display_name: puptoo pr-check
      - "insights-gh-build-master":
          display_name: puptoo build-master

Gitlab:

.. code-block:: yaml

    project:
      name: insightsapp-poc-ci
      label: insights
      node: insights
      gl_group: bsquizza
      gl_project: insights-ingress-go
      quay_org: cloudservices
      jobs:
      - 'insights-gl-pr-check':
          display_name: 'insightsapp-poc pr-check'
      - 'insights-gl-build-master':
          display_name: 'insightsapp-poc build-master'


In your app's build.yml, you need to specify on which Jenkins server to have
your jobs defined.  App SRE provides two Jenkins servers: ci-int for projects
hosted on gitlab.cee.redhat.com, and ci-ext for public projects hosted on
Github.  Note that private Github projects are **not supported**; if a Github
project must remain private, then its origin must move to gitlab.cee.redhat.com.

Create deployment template with ClowdApp resource
-------------------------------------------------

Going forward, an app's deployment template must live in its source code repo.
This will simply saas-deploy file configuration (see below) and has always been
App SRE's convention.

Additional resources defined in an app's current deployment template besides
Deployment and Service should be copied over to the new template in the app's
source code repo.  Then the ClowdApp developed above should be added in.

A ClowdApp must point to a ClowdEnvironment resource via its ``envName`` spec
attribute, and its value should be set as the ``ENV_NAME`` template parameter.

Modify saas-deploy file for service
-----------------------------------

* Github projects need to create a separate saas-deploy file because it needs
  to point to ci-ext
* Add ClowdApp as a resource type
* Point resource template URL and path to deployment template in code repo
* Remove IMAGE_TAG from all targets
* Ensure ref is set to master for stage and a git SHA for production.
* Add ephemeral target

Disable builds in e2e-deploy
----------------------------

* Remove BuildConfig resources from buildfactory folder.
* Provide example PR

.. vim: tw=80 spelllang=en
