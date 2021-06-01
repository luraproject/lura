# Contributing

Thank you for your interest in contributing to Lura, there are several ways
you can contribute and make this project more awesome, please see below:

## Reporting an Issue

If you believe you have found an issue with the code please do not hesitate to file an issue in [Github](https://github.com/luraproject/lura/issues). When
filing the issue please describe the problem with the maximum level of detail
and the steps to reproduce the problem, including information about your
environment.

You can also open an issue requesting for help or doing a question and it's
also a good way of contributing since other users might be in a similar
position.

Please note we have a code of conduct, please follow it in all your interactions with the project.

## Code Contributions

When contributing to this repository, it is generally a good idea to discuss
the change with the owners before investing a lot of time coding. The process
could be:

1. Open an issue explaining the improvment or fix you want to add
2. [Fork the project](https://github.com/luraproject/lura/fork_select)
3. Code it in your fork
4. Submit a [pull request](https://help.github.com/articles/creating-a-pull-request) referencing the issue


Your work will then be reviewed as soon as possible (suggestions about some
changes, improvements or alternatives may be given).

**Don't forget to add tests**, make sure that they all pass!

# Help with Git

Once the repository is forked, you should track the upstream (original) one
using the following command:

    git remote add upstream https://github.com/luraproject/lura.git

Then you should create your own branch:

    git checkout -b <prefix>/<micro-title>-<issue-number>

Once your changes are done (`git commit -am '<descriptive-message>'`), get the
upstream changes:

    git checkout master
    git pull --rebase origin master
    git pull --rebase upstream master
    git checkout <your-branch>
    git rebase master

Finally, publish your changes:

    git push -f origin <your-branch>

You should be now ready to make a pull request.
