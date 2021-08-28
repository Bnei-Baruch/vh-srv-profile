#!/bin/bash

export DB_HOST="profiledb"
export DB_PORT=5436
export DB_USER="dev_db_user"
export DB_DATABASE="dev_profile_db"
export DB_PASSWORD="password"
export PGPASSWORD="password"
export DATABASE_URL="postgres://${DB_USER}:password@localhost:${DB_PORT}/${DB_DATABASE}"
export APP_PORT=":7471"
export APP_MODE="dev"
export IMAGE_NAME="vh-srv-profile"
export CORSACTIVE="true"

case $1 in 
	"dbup")
		docker network create vh
		sudo [ ! -d /opt/srv-profile-db ] && mkdir /opt/srv-profile-db
		sudo cp -f db/initial.sql /opt/srv-profile-db/initial.sql 
		docker-compose -f docker-compose.local.yml up -d
		;;
	"dbdown")
 	docker-compose -f docker-compose.local.yml down
	;;
	"dbconnect")
		psql -h localhost --port $DB_PORT -d $DB_DATABASE -U $DB_USER
	;;
	"dbexec")
		psql -h localhost --port $DB_PORT -d $DB_DATABASE -U $DB_USER -c "$2"
	;;
	"dbrun")
		psql -h localhost --port $DB_PORT -d $DB_DATABASE -U $DB_USER < "$2"
	;;
	"build")
	go build .
	;;
    "run")
	./vh-srv-profile
esac
