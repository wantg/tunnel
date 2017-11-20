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
        srcHost: '127.0.0.1',
        srcPort: 27017,
        dstHost: '127.0.0.1',
        dstPort: 27017
    },
    redis: {
        sshConfig: {
            host: '',
            port: 22,
            username: '',
            password: ''
        },
        srcHost: '127.0.0.1',
        srcPort: 6379,
        dstHost: '127.0.0.1',
        dstPort: 6379
    }
};

function startService(forwardConfig) {
    net.createServer((netConnection) => {
        const sshConnection = new sshClient();
        sshConnection.on('ready', () => {
            sshConnection.forwardOut(
                forwardConfig.srcHost,
                forwardConfig.srcPort,
                forwardConfig.dstHost,
                forwardConfig.dstPort,
                (err, sshStream) => {
                    if (err) {
                        throw err;
                    }
                    netConnection.pipe(sshStream).pipe(netConnection);
                }
            );
        }).connect(forwardConfig.sshConfig);
    }).listen(forwardConfig.srcPort, forwardConfig.srcHost, (error, result) => {
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