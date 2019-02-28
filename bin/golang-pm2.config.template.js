/*
 Country string 国家
 AreaCode int 区域
 LEVEL string
 VpnMgrCore string 中心服务器地址 ip:port
 COUNTER_URL string 统计后台地址 ip:port
 */
module.exports = {
    apps: [
        {
            name: "IggAgent_Country_AreaCode_LEVEL",
            cwd: "IggAgent",
            script: "IggAgent_linux",
            args: ['-url', 'COUNTER_URL', '-size', 100,],
            error_file: "/data/vpnedge/logs/IggAgent_Country_LEVEL.err.log",
            out_file: "/data/vpnedge/logs/IggAgent_Country_LEVEL.out.log",
        },
        {
            name: "ssEdge_Country_AreaCode_LEVEL_1",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', 'VpnMgrCore', '-Id', 1, '-beginport', 20000, '-endport', 29999,
                '-state', 'Country', '-area', 'AreaCode'],
            error_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_1.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_1.out.log",
        },
        {
            name: "ssEdge_Country_AreaCode_LEVEL_2",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', 'VpnMgrCore', '-Id', 2, '-beginport', 30000, '-endport', 39999,
                '-state', 'Country', '-area', 'AreaCode'],
            error_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_2.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_2.out.log",
        },
        {
            name: "ssEdge_Country_AreaCode_LEVEL_3",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', 'VpnMgrCore', '-Id', 3, '-beginport', 40000, '-endport', 49999,
                '-state', 'Country', '-area', 'AreaCode'],
            error_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_3.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_3.out.log",
        },
        {
            name: "ssEdge_Country_AreaCode_LEVEL_4",
            cwd: "VpnApp",
            script: "vpnedge_linux",
            args: ['-url', 'VpnMgrCore', '-Id', 4, '-beginport', 50000, '-endport', 59999,
                '-state', 'Country', '-area', 'AreaCode'],
            error_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_4.err.log",
            out_file: "/data/vpnedge/logs/ssEdge_Country_AreaCode_LEVEL_4.out.log",
        }
    ]
};

