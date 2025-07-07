## install

run `yarn install` to install the dependencies

## build

run `yarn build` to build the project


## Notes for Users Behind a  Proxy
If you are behind a  proxy, running build-frontend.bat may fail due to Yarn not automatically using the http_proxy and https_proxy environment variables.

To address this, the build-frontend.bat script has been enhanced to help configure Yarn's proxy settings interactively.
During execution, the script will ask whether you are behind a proxy.

If you confirm, it will attempt to read your system's http_proxy and https_proxy environment variables as default values.

You can accept the defaults or input custom proxy addresses. It then runs:

`yarn config set httpProxy <your_proxy>`

`yarn config set httpsProxy <your_proxy>`

to properly configure Yarn.
After this setup, the script will continue with the build process.
