# Skadi golang SDK
[hack-fan/skadi](https://github.com/hack-fan/skadi) is a cloud message center,
you can send your message/job/command to it by several ways, Slack/Teams/Wechat etc...
then your agent will get it, do anything you defined, return the result.

This is golang sdk for hack-fan/skadi, pull your jobs from skadi server.

## Worker
Worker is a daemon thread in your app, fetch message every minute and handle it.

## Reporter
Report can send info/warning to your pre-defined IM.
