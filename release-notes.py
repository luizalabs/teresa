# -*- coding: utf-8 -*-
import re

with open('CHANGELOG.md') as f:
    content = f.read()

matches = re.search(r'## \[(.|\n)*?## \[', content)
print '\n'.join(matches.group().splitlines()[1:-1])
