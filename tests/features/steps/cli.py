import os

package = "sh"
name = os.environ["cli"]
platform = getattr(__import__(package, fromlist=[name]), name)

@when('we run the {text} command')
def cli_command(context, text):
    context.response=platform(text)