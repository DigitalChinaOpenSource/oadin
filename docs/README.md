## Build Steps

Ensure Graphviz is installed beforehand.

```sh

pip install sphinx
pip install sphinx-rtd-theme


cd docs

# build Chinese version
cd zh-cn  && make html

# build English version
cd en && make html
```
