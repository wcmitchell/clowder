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

* Use env var to switch between consuming configuration from clowder and from
  current config method (e.g. env vars, ConfigMap)
* Dependent service hostnames
* Kafka bootstrap URL
* Kafka topic names
* Web prefix and port number
* Metrics path and port number
* Use minio as the only object storage client library
* Redis

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
--------------------------------------------------------

* build_deploy.sh should be cloned from example
* Builds image using Dockerfile and pushes to quay with credentials provided in
  Jenkins job environment
* Push latest and qa tags for e2e-deploy backwards compatibility
* Clone pr_check.sh from example and fill in bonfire variables
* Both files live in the root folder of source code repo

Create PR check and build_master jenkins jobs in app-interface
--------------------------------------------------------------

* Copy from template and fill in the blanks
* Public github repos should use ci-ext
* Private repos should live on gitlab and use ci-int
* Private github repos are not supported

Create deployment template with ClowdApp resource
-------------------------------------------------

* Standard parameter ENV_NAME
* Simply copy in ClowdApp developed above

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
