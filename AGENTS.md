## working agreements

* when asked to do something, go into a planning mode, and crete a detailed plan of proposed changes. Check with me first before making any code changes, and ensure that I am in agreement with the plan before moving forward. We may need to work on it first before starting code changes. If I ask for changes to the plan and you make them, do not then assume the plan is approved, re ask me to review the revised plan.
* Do work in small increments that can easily be rolled back, after each increment check in with me to ensure I am aligned with the changes.
* If I give you a the name or path of a file or directory, make sure to actually read the contents fully and tie them in to the planning.
* read the contents of the project README.md @README.md, internalize the developer guide before starting work or planning.
* if you attempt to run a command and are blocked due to sandbox restrictions, you can elevate to me and request approval to run the command.
* you must avoid changing the state of any system other than the local dev environment or local docker containers.
* you must never attempt to run an AWS, AZURE, or GCP command.
* you must never run git push.
* when asked to do a task, explore the code base for prior examples of similar tasks, e.g. implementing an interface, or writing tests. You should follow established patterns in the codebase for new work.
* after making changes trace the changes and ensure they are sound logicially, use tools such as building modules and running tests to aid in validation.
* you should favor a test driven development workflow where you can validate your small incremental changes early before making more code changes, to ensure that you are not adding too much code that may need large refactors later because they do not build properly or fail tests.
