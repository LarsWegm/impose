
<div align="center">
  <img src="./logo.png" width="200" />
  
  # Image version updater for Docker Compose
  
  A CLI tool to automatically update image versions in Docker Compose files.
  <br/><br/>
  
</div>

Using the `latest` tag for production Docker services is not recommended (see e.g. https://vsupalov.com/docker-latest-tag/). But always manually looking for the latest version tags and updating them in your Docker Compose file is tedious. That's where this Command Line tool comes in.

The tool checks your `docker-compose.yml` and updates all Docker images to the latest tag. You can even choose if you want to update based on major, minor or patch versions. The tool can also warn you on specific version changes or ignore some services at all.

## Quick Start

It's really simple. Go to the directory of your `docker-compose.yml` file and run `impose update` (attention, this will overwrite your current `docker-compose.yml`).

You should see an output similar to this:

```sh
Changed versions:
  alpine:3.15.5 => alpine:3.16.3
```

All outdated image tags should now be updated to the latest version:

```diff
--- a/docker-compose.yml
+++ b/docker-compose.yml
@@ -1,4 +1,4 @@
 version: '3'
 services:
     my-service:
-        image: alpine:3.15.5
+        image: alpine:3.16.3
```

## Usage

You can use head or inline comments for the image keyword in the Docker Compose file to add annotations.

The following annotations are available:

```
  impose:ignore     ignores the image for updates
  impose:minor      only checks for minor version updates
  impose:patch      only checks for patch version updates
  impose:warnMajor  warns if major version has changed
  impose:warnMinor  warns if minor version has changed (including major version changes)
  impose:warnPatch  warns if patch version has changed (including major and minor version changes)
  impose:warnAll    warns if the version string has changed in any way (including version suffix)
```

For example, you can apply the annotations as follows:

```yaml
version: '3'
services:
    my-service:
        image: alpine:3.15.5 # impose:minor
```

Use the `--help` flag for more information about the commands and options.

## Development

This CLI program uses the [Cobra](https://github.com/spf13/cobra) Go library together with the corresponding scaffolding tool [Cobra CLI](https://github.com/spf13/cobra-cli).
