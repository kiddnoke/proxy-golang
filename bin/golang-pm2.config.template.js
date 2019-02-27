module.exports = {
    apps: [
        {
            name: "IggAgent_COUNTRY",
            cwd: "IggAgent",
            script: "IggAgent_linux",
            args: ['-url', 'COUNTER_URL', '-size', 100,],
            error_file: "/data/vpnedge/logs/IggAgent_COUNTRY.err.log",
            out_file: "/data/vpnedge/logs/IggAgent_COUNTRY.out.log",
        },
        {
            name: "ssEdge_COUNTRY_1",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', '47.74.157.30:7001', '-Id', 1, '-beginport', 20000, '-endport', 20999,
                '-state', 'COUNTRY', '-area', '1'],
            error_file: "/data/vpnedge/logs/ssEdge_COUNTRY_1.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_COUNTRY_1.out.log",
        },
        {
            name: "ssEdge_COUNTRY_2",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', '47.74.157.30:7001', '-Id', 2, '-beginport', 21000, '-endport', 21999,
                '-state', 'COUNTRY', '-area', '2'],
            error_file: "/data/vpnedge/logs/ssEdge_COUNTRY_2.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_COUNTRY_2.out.log",
        },
        {
            name: "ssEdge_COUNTRY_3",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', '47.74.157.30:7001', '-Id', 3, '-beginport', 22000, '-endport', 22999,
                '-state', 'COUNTRY', '-area', '3'],
            error_file: "/data/vpnedge/logs/ssEdge_COUNTRY_3.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_COUNTRY_3.out.log",
        },
        {
            name: "ssEdge_COUNTRY_4",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', '47.74.157.30:7001', '-Id', 4, '-beginport', 23000, '-endport', 23999,
                '-state', 'COUNTRY', '-area', '4'],
            error_file: "/data/vpnedge/logs/ssEdge_COUNTRY_4.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_COUNTRY_4.out.log",
        }
    ]
};

