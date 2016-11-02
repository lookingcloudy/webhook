
# Bitbucket Webhook
This is a fork of [adnanh's very fine webhook](https://github.com/adnanh/webhook/) module and customized specifically for Bitbucket.  

Bitbucket webhooks contain an array of changes, which may include  branch or tag changes.  This module is specifically design to trigger on either a branch or tag name in any of the pushed changes.   A branch or tag can be matched as an exact value, or as a regex pattern.

Much of the original functionality is not needed for this simple application and has been stripped out.

The branch or tag name that was matched is passed to the command as the first and only parameter.

Here is a sample `hooks.json` file:
```json
[
  {
    "id": "developBranch",
    "execute-command": "/root/deploy-develop.sh",
    "command-working-directory": "/root",
    "response-message": "I got the payload!",
    "trigger-rule":
    {
      "type": "value",
      "source": "branch",
      "value": "develop"
    }
  },
  {
    "id": "qa-builds",
    "execute-command": "/root/deploy-qa.sh",
    "command-working-directory": "/root",
    "response-message": "I got the payload!",
    "trigger-rule":
    {
      "match": {
          "type": "regex",
          "source": "tag",
          "value": "^.*-qa$"
      }
    }
  }
]
```


On Ubuntu, this is a sample `webhook.conf` init file.  Place this in `/etc/init/webhook.conf`:
```
description "webhooks"

start on (filesystem and net-device-up IFACE!=lo)
stop on runlevel [!2345]

respawn

kill timeout 20

script

        exec /root/webhook/webhook -verbose -hooks=/root/webhook/hooks.json

end script
```


And here is a sample build script for a go application.  This assumes that a Dockerfile exists in the root folder.  The webhook application passes the matched parameter of the branch or tag to the script.  In the case of a tag match the matched value is prefixed by "tags/".  By doing this a single build script can be used for both build on branches or tags.  In the build script the "tags/" prefix is stripped off and stored in $TAG.  In this build script, NGINX is used to proxy the application so the port is not exposed.
```
#!/bin/bash


TAG=${1//tags\//} 


# DEVELOPMENT BUILD
DOCKER_NAME="dev-appname"
DOCKER_TAG="myname/appname:$TAG"
GIT_URL="git@bitbucket.org/myaccount/appname.git"
REPO="appname"

echo "Getting currently running containers"
OLDPORTS=( `docker ps -a | grep $DOCKER_NAME | awk '{print $1}'`)


echo "removing old source folder"
if [ -d floodquote ]; then
  rm -rf floodquote
fi

echo "pulling new version"
git clone $GIT_URL
cd $REPO
git checkout -f $1

docker build --no-cache=true -t $DOCKER_TAG .

echo "removing old containers"
for i in ${OLDPORTS[@]}
do
        echo "removing old container $i"
        docker kill $i
        docker rm $i
done

sleep 5

echo "starting new containers"
docker run --restart=always --network=private -e APPVER=$TAG -d --name $DOCKER_NAME $DOCKER_TAG

```
