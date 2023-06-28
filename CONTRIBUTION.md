# Contribution Guide

Thanks for taking out time to read our contribution guide. We are excited to see
what to have in mind for B.O.B.

## Community Guidelines

We want to keep this project awesome, growing and collaborative. We need your
help to keep it that way. To help with this we've come up with some general
guidelines for the community as a whole:
- Be nice: Be courteous, respectful and polite to fellow community members: no
  regional, racial, gender, or other abuse will be tolerated. We like nice
  people way better than mean ones!
- Encourage diversity and participation: Make everyone in our community feel
  welcome, regardless of their background and the extent of their contributions,
  and do everything possible to encourage participation in our community.
- Keep it legal: Basically, don't get us in trouble. Share only content that you
  own, do not share private or sensitive information, and don't break the law.
- Stay on topic: Make sure that you are posting to the correct channel and avoid
  off-topic discussions. Remember when you update an issue or respond to an
  email you are potentially sending to a large number of people. Please consider
  this before you update. Also remember that nobody likes spam.
- Don't send email to the maintainers: There's no need to send email to the
  maintainers to ask them to investigate an issue or to take a look at a pull
  request. Instead of sending an email, GitHub mentions should be used to ping
  maintainers to review a pull request, a proposal or an issue.

## Coding Style

Unless explicitly stated, we follow all coding guidelines from the Go community.
While some of these standards may seem arbitrary, they somehow seem to result in
a solid, consistent codebase.

It is possible that the code base does not currently comply with these
guidelines. We are not looking for a massive PR that fixes this, since that goes
against the spirit of the guidelines. All new contributions should make a best
effort to clean up and make the code base better than they left it. Obviously,
apply your best judgement. Remember, the goal here is to make the code base
easier for humans to navigate and understand. Always keep that in mind when
nudging others to comply.

The rules:

- All code should be formatted with gofmt -s.
- All code should pass the default levels of golint.
- All code should follow the guidelines covered in [Effective
  Go](http://golang.org/doc/effective_go.html) and [Go Code Review
  Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Comment the code. Tell us the why, the history and the context.
- Document all declarations and methods, even private ones. Declare
  expectations, caveats and anything else that may be important. If a type gets
  exported, having the comments already there will ensure it's ready.
- Variable name length should be proportional to it's context and no longer.
  noCommaALongVariableNameLikeThisIsNotMoreClearWhenASimpleCommentWouldDo. In
  practice, short methods will have short variable names and globals will have
  longer names.
- No underscores in package names. If you need a compound name, step back, and
  re-examine why you need a compound name. If you still think you need a
  compound name, lose the underscore.
- No utils or helpers packages. If a function is not general enough to warrant
  it's own package, it has not been written generally enough to be a part of a
  util package. Just leave it unexported and well-documented.
- All tests should run with go test and outside tooling should not be required.
  No, we don't need another unit testing framework. Assertion packages are
  acceptable if they provide real incremental value.
- Even though we call these "rules" above, they are actually just guidelines.
  Since you've read all the rules, you now know that.

If you are having trouble getting into the mood of idiomatic Go, we recommend
reading through [Effective Go](http://golang.org/doc/effective_go.html). The [Go
Blog](https://blog.golang.org/) is also a great resource.