@then('we should see "{text}"')
def should_see(context, text):
    if text not in context.response:
        raise Exception('\n\n%r not in %r' % (text, context.response))

@then('we should not see "{text}"')
def should_not_see(context, text):
    if text in context.response:
        raise Exception('\n\n%r not in %r' % (text, context.response))
