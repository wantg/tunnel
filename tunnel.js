'use strict';

const net = require('net');
const sshClient = require('ssh2').Client;

const forwardConfigs = {
    mongodb: {
        sshConfig: {
            host: '',
            port: 22,
            username: '',
            password: ''
        },
        src: { host: '127.0.0.1', port: 27017, },
        dst: { host: '127.0.0.1', port: 27017 }
    },
    redis: {
        sshConfig: {
            host: '',
            port: 22,
            username: '',
            password: ''
        },
        src: { host: '127.0.0.1', port: 6379, },
        dst: { host: '127.0.0.1', port: 6379 }
    }
};

function startService(forwardConfig) {
    net.createServer((netConnection) => {
        const sshConnection = new sshClient();
        sshConnection.on('ready', () => {
            sshConnection.forwardOut(
                forwardConfig.src.host,
                forwardConfig.src.port,
                forwardConfig.dst.host,
                forwardConfig.dst.port,
                (err, sshStream) => {
                    if (err) {
                        throw err;
                    }
                    netConnection.pipe(sshStream).pipe(netConnection);
                }
            );
        }).connect(forwardConfig.sshConfig);
    }).listen(forwardConfig.src.port, forwardConfig.src.host, (error, result) => {
        if (error) {
            console.log(error);
        } else {
            console.log('success');
        }
    });
}

for (let title in forwardConfigs) {
    console.log(title + ' start');
    startService(forwardConfigs[title]);
}