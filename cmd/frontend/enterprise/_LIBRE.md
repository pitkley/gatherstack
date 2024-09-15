> Why does this folder still exist?

The code here, as-is, creates a lot of dummy handlers and Sourcegraph injects the real handlers in
actual enterprise code. Removing it requires refactoring throughout, which was not  done yet.

> Does this violate any license?

It should not, because the README as of the last OSS commit
(`1cd36d2dbbd2a9ab638cc437d208d2717eaefb0b`) states that only the root `enterprise/` folder and the
`client/web/src/enterprise/` folder fell under the non-OSS enterprise license.
