# Slurmer

A job manager for Slurm, written in Go.

## Installation

### From sources

First, you have to compile the code to get both `slurmer` and `executor` executables. A Makefile is available to do so:

```bash
make
```

There is few variables that you can override to specify custom options:

| Variable name          | Description                                              | Default      |
| ---------------------- | -------------------------------------------------------- | ------------ |
| _MIN\_UID_, _MAX\_UID_ | the minimum and maximum uids that the executor can spoof | 1000, 65565  |
| _SLURMER\_UID_         | The user that is authorized to use the executor          | 1000         |
| PREFIX                 | The path prefix to install slurmer to                    | `/usr/local` |

Don't forget to `make clean` between variable changes to make sure that your changes have taken effect.

#### Example of custom variables

```bash
make MIN_UID=1100 MAX_UID=1200 SLURMER_UID=1099
```

Then you can install the executables to the destination directory (need root permissions):

```bash
make PREFIX=/opt/slurmer install
```

Now Slurmer should be installed on your system. Next we'll see how to use it.

## Configuration

For this section, let's consider that slurmer is installed to prefix `/usr/local/`. Let's say that we have the configuration in `/usr/local/etc/slurmer/config.yml` (see [sample configuration](#sample-configuration)) and the Slurmer files have to be in `/var/lib/slurmer`. Make sure that only Slurmer account can edit `config.yml`.

First, create a directory `/var/lib/slurmer` and set the owner to the Slurmer user (we assume that `slurmer` is the Slurmer user):

```bash
mkdir /var/lib/slurmer
chown -R slurmer /var/lib/slurmer/ /usr/local/etc/slurmer/

# make sure that only slurmer can access and modify these files
chmod 0700 /var/lib/slurmer/ /usr/local/etc/slurmer/
chmod 0600 /usr/local/etc/slurmer/config.yml
```

Next, set the value `working_dir` in `config.yml` to `/var/lib/slurmer`.

Now you can run Slurmer with the following line to consider the given configuration (run with the Slurmer user):

```bash
slurmer -c /usr/local/etc/slurmer/config.yml
```

### Sample configuration

A sample config file is also available in `configs/config-example.yml`.

```yaml
# Useful when 'connector' is set to 'slurmrest'; not implemented yet
slurmrest:
  url: "http+unix:///default?socket=/tmp/slurmrestd.sock"

slurmer:
  ip: 127.0.0.1
  port: 8080
  connector: slurmcli
  working_dir: /var/lib/slurmer
  templates_dir: /etc/slurmer/templates
  executor_path: /usr/bin/executor
  applications:
    - name: app1
      token: averycomplicatedchallenge
      uuid: a9a5fc66-9bed-4a13-874a-d8d7d1756224 # a random app uuid
  logs:
    # available: text, json
    format: text

    stdout: false
    output: /var/logs/slurmer.log

    # available, from most to less verbose:
    # trace, debug, info, warning, error, fatal, panic
    level: debug

# To enable OIDC
oidc:
  enabled: false
  issuer: ""
  audience: ""
```

## REST API summary

Each request must be authenticated with an application token and eventually a user for jobs related resources.

| Header       | Description                  |
| ------------ | ---------------------------- |
| X-Auth-Token | Authenticates an application |
| User         | The user that owns a job     |

### Routes

| Verb   | Route                             | Description                                       |
| ------ | --------------------------------- | ------------------------------------------------- |
| GET    | /apps                             | list all applications, need admin access token    |
| POST   | /apps                             | add a new application, need admin access token    |
| GET    | /apps/{appId}                     | show {appId} application, need admin access token |
| GET    | /apps/{appId}/jobs                | list all jobs of the app {appId}                  |
| POST   | /apps/{appId}/jobs                | add a new job to app {appId}                      |
| GET    | /apps/{appId}/jobs/{jobId}        | show job {jobId} details                          |
| DELETE | /apps/{appId}/jobs/{jobId}        | delete a registered job                           |
| GET    | /apps/{appId}/jobs/{jobId}/batch  | read jobs batch file                              |
| GET    | /apps/{appId}/jobs/{jobId}/out    | read jobs standard output                         |
| GET    | /apps/{appId}/jobs/{jobId}/files  | download the job files as zip                     |
| POST   | /apps/{appId}/jobs/{jobId}/files  | upload a zip to job directory                     |
| PUT    | /apps/{appId}/jobs/{jobId}/status | start or stop a job                               |
| PUT    | /apps/{appId}/jobs/{jobId}/prune  | clear all job data except batch file              |
| PUT    | /apps/{appId}/jobs/{jobId}/start  | start a job                                       |
| PUT    | /apps/{appId}/jobs/{jobId}/stop   | stop a job                                        |
