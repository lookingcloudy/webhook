
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
      "type": "regex",
      "source": "tag",
      "value": "^.*-qa$"
    }
  }
]
```


