docker build -t registry.heroku.com/panch-foostrack/web .
docker push registry.heroku.com/panch-foostrack/web
heroku container:release web -a panch-foostrack