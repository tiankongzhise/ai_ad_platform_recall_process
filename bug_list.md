1.查询逻辑存在明显的问题，https://ad-platform-recall.hnzzzsw.com/api/query?recall_service_name=test&platform=1&user_name=2&page=1&page_size=20返回了完全不对的数据，{
    "code": 0,
    "message": "查询成功",
    "data": {
        "total": 1,
        "page": 1,
        "page_size": 20,
        "records": [
            {
                "id": 1,
                "recall_service_name": "admin",
                "platform": "1",
                "user_name": "1",
                "params": "{\"app_id\":\"1863717820933212\",\"auth_code\":\"a962b75b4ee19349e3b6bce86c407ad86819f4ef\",\"material_auth_status\":\"0\",\"platform_number\":\"1\",\"recall_service_user_uid\":\"de7482a730210cffa154c69c97873697\",\"scope\":\"[200000032,300000000,130,4,300000006,300000040,300000041,110,112,200000018,300000052,14,120,122,123,124,300000029]\",\"uid\":\"92403628013\",\"user_number\":\"1\"}",
                "created_at": "2026-04-28T22:34:48.341+08:00"
            }
        ]
    }
}
2.用户注销以后，如果新注册的用户与注销用户同名，当前查询逻辑下，新用户能查询老用户的历史数据。这是不对的。需要添加UID校验，确保新注册用户不能查询注销老用户数据。