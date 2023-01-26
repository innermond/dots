First of all start the container.
`docker compose up`

#create a migration#
Do not forget to quote (with duble quoyes ") table names that are words used by postgres, like user.

`migrate create -ext sql -dir migrations -seq -digits 3 <migration name>`

##run migration onto dockerized postgresql
`migrate -path migrations -database "postgresql://<user>:<password>@localhost:5432/dots?sslmode=disable" -verbose up`

On development user is postgres and password is admin.

The postgres database from docker must be connected, which may be bone connecting to docker.

1 `docker container exec -it <docker-name> /bin/bash`
docker-name may be dots-db-1.
2 Once inside docker `su postgres`
3 psql
4 \c dots

##create/update .securecookie
This file *.securecookie* it is used by gorilla securecookie. It contains 2 lines, for hashKey and blockKey 

###create a new one
`openssl rand -hex 64 > .securecookie && openssl rand -hex 32 >> .securecookie`
