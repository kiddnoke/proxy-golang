module.exports = {
    apps: [{
        name: "ssEdge_SG_1",
        cwd: '/home/zhangsen',
        script: "vpnedge_linux",
        args: ['-url', '10.0.2.70:7001', '-Id', 1, '-beginport', 20000, '-endport', 20999, '-state', 'SG', '-area', '1'],
    }, {
        name: "ssEdge_SG_2",
        cwd: '/home/zhangsen',
        script: "vpnedge_linux",
        args: ['-url', '10.0.2.70:7001', '-Id', 2, '-beginport', 21000, '-endport', 21999, '-state', 'SG', '-area', '2'],
    }, {
        name: "ssEdge_GE_1",
        cwd: '/home/zhangsen',
        script: "vpnedge_linux",
        args: ['-url', '10.0.2.70:7001', '-Id', 3, '-beginport', 22000, '-endport', 22999, '-state', 'GE', '-area', '1'],
    }, {
        name: "ssEdge_GE_2",
        cwd: '/home/zhangsen',
        script: "vpnedge_linux",
        args: ['-url', '10.0.2.70:7001', '-Id', 4, '-beginport', 23000, '-endport', 23999, '-state', 'GE', '-area', '2'],
    }, {
        name: "ssEdge_HK_1",
        cwd: '/home/zhangsen',
        script: "vpnedge_linux",
        args: ['-url', '10.0.2.70:7001', '-Id', 5, '-beginport', 24000, '-endport', 24999, '-state', 'HK', '-area', '1'],
    }, {
        name: "ssEdge_HK_2",
        cwd: '/home/zhangsen',
        script: "vpnedge_linux",
        args: ['-url', '10.0.2.70:7001', '-Id', 6, '-beginport', 25000, '-endport', 25999, '-state', 'HK', '-area', '2'],
    }, {
        name: "IggAgent",
        script: "IggAgent",
        args: ['-url', '10.0.3.234:6061', '-size', 30,],
    },]
};
