#ifndef ROC_APP
#define ROC_APP

#include "roc_std.h"

struct Header {
    struct RocStr name;
    struct RocStr value;
};

struct Request {
    struct RocList body;
    struct RocList headers;
    struct RocStr url;
    unsigned char methodEnum;
};

struct Response {
    struct RocList body;
    struct RocList headers;
    short unsigned int status;
};

union ResultModelUnion {
    struct RocStr error;
    void* *model;
};

struct ResultModel {
    union ResultModelUnion payload;
    unsigned char disciminant;
};


union ResultResponseUnion {
    struct RocStr error;
    struct Response response;
};

struct ResultResponse {
    union ResultResponseUnion payload;
    unsigned char disciminant;
};

extern void roc__init_model_for_host_1_exposed_generic(void* *model);
extern void roc__update_model_for_host_1_exposed_generic(const struct ResultModel *new_model, void* model, struct RocList *events);
extern void roc__handle_request_for_host_1_exposed_generic(const struct ResultResponse *response, struct Request *request, void* model);

#endif
