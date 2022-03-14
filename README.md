# Slurmer
A job manager for Slurm, written in Go

## Headers

| Header       | Description                  |
| ------------ | ---------------------------- |
| X-Auth-Token | Authenticates an application |

```rest
GET     /jobs               # list all jobs of the current app
GET     /jobs/{id}          # show job {id} details
POST    /jobs               # add a new job
PUT     /jobs/{id}          # edit a job
PATCH   /jobs/{id}          # edit parts of a job
PUT     /jobs/{id}/status   # start or stop a job
DELETE  /jobs/{id}          # delete a registered job
```