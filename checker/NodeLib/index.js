// const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawn } = require('child_process');

class Byze {
  IsByzeAvailable(){
      return new Promise((resolve) => {
          const options = {
            hostname: 'localhost',
            port: 16688,
            path: '/',
            method: 'GET',
            timeout: 3000,
          };
          const req = http.request(options, (res) => {
            resolve(res.statusCode === 200);
          });
          req.on('error', () => resolve(false));
          req.on('timeout', () => {
            req.destroy();
            resolve(false);
          });
          req.end();
      });
  }

  IsByzeExist(){
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'Byze');
      const dest = path.join(destDir, 'Byze.exe');
      return fs.existsSync(dest);
  }

  DownloadByze() {
      return new Promise((resolve) => {
        const url = 'http://120.232.136.73:31619/Byzedev/Byze.exe';
        const userDir = os.homedir();
        const destDir = path.join(userDir, 'Byze');
        const dest = path.join(destDir, 'Byze.exe');
        fs.mkdir(destDir, { recursive: true }, (err) => {
          if (err) return resolve(false);
          const file = fs.createWriteStream(dest);
          http.get(url, (res) => {
            res.pipe(file);
            file.on('finish', () => {
              file.close(() => resolve(true));
            });
          }).on('error', () => resolve(false));
        });
      });
  }

  InstallByze() {
    return new Promise((resolve) => {
      const userDir = os.homedir();
      const byzePath = path.join(userDir, 'Byze', 'byze.exe');
      const child = spawn(byzePath, ['server', 'start', '-d'], { detached: true, stdio: 'ignore' });
  
      child.on('error', () => resolve(false));
      setTimeout(() => resolve(true), 3000); // wait 3s for Byze server to start
      child.unref();
    });
  }

  ByzeInit(){

  }
}

module.exports = Byze;