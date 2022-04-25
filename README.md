# Slurmer
A job manager for Slurm, written in Go

## Headers

| Header       | Description                  |
| ------------ | ---------------------------- |
| X-Auth-Token | Authenticates an application |

```rest
GET     /apps                               # list all applications, debug infos
GET     /apps/{appId}                       # show {appId} application, debug info
GET     /apps/{appId}/jobs                  # list all jobs of the app {appId}
POST    /apps/{appId}/jobs                  # add a new job to app {appId}
GET     /apps/{appId}/jobs/{jobId}          # show job {jobId} details
DELETE  /apps/{appId}/jobs/{jobId}          # delete a registered job, unimplemented yet
PUT     /apps/{appId}/jobs/{jobId}/status   # start or stop a job
GET     /apps/{appId}/jobs/{jobId}/files    # download the job files as zip
POST     /apps/{appId}/jobs/{jobId}/files    # upload a zip to job directory
```