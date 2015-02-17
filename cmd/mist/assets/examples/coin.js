var walletABI = [
    {
        "name":"confirm",
        "type":"function",
        "constant":false,
        "inputs":[
            {"name":"_h","type":"hash256"}
        ],
        "outputs":[]
    },{
        "name":"execute",
        "constant":false,
        "type":"function",
        "inputs":[
            {"name":"_to","type":"address"},
            {"name":"_value","type":"uint256"},
            {"name":"_data","type":"bytes"}
        ],
        "outputs":[
            {"name":"_r","type":"hash256"}
        ]
    },{
        "name":"kill",
        "type":"function",
        "constant":false,
        "inputs":[
            {"name":"_to","type":"address"}
        ],
        "outputs":[]
    },{
        "name":"changeOwner",
        "type":"function",
        "constant":false,
        "inputs":[
            {"name":"_from","type":"address"},
            {"name":"_to","type":"address"}
        ],
        "outputs":[]
    },{
        "name":"CashIn",
        "type":"event",
        "inputs":[
            {"indexed":false,"name":"value","type":"uint256"}
        ]
    },{
        "name":"SingleTransact",
        "type":"event",
        "inputs":[
            {"indexed":true,"name":"out","type":"string32"},
            {"indexed":false,"name":"owner","type":"address"},
            {"indexed":false,"name":"value","type":"uint256"},
            {"indexed":false,"name":"to","type":"address"}
        ]
    },{
        "name":"MultiTransact",
        "type":"event",
        "inputs":[
            {"indexed":true,"name":"out","type":"string32"},
            {"indexed":false,"name":"owner","type":"address"},
            {"indexed":false,"name":"operation","type":"hash256"},
            {"indexed":false,"name":"value","type":"uint256"},
            {"indexed":false,"name":"to","type":"address"}
        ]
    }
];