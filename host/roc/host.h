#include <stdlib.h>

struct RocStr {
    char* bytes;
    size_t len;
    size_t capacity;
};

struct RocList {
    char* bytes;
    size_t len;
    size_t capacity;
};

union RequestTimeoutUnion {
    long long unsigned int timeoutMilliseconds;
};

struct RequestTimeout {
    union RequestTimeoutUnion payload;
    unsigned char discriminant;
};

struct Header {
    struct RocStr name;
    struct RocList value;
};

struct Request {
    struct RocList body;
    struct RocList headers;
    struct RocStr mimeType;
    struct RequestTimeout timeout;
    struct RocStr url;
    unsigned char methodEnum; 
};

struct Response {
    struct RocStr body;
    struct RocList headers;
    short unsigned int status;
};

struct ResponseModel {
    struct Response response;
    void* *model;
};

union DecodeArgUnion {
    struct RocList bytes;
};

struct DecodeArg {
    union DecodeArgUnion payload;
    unsigned char discriminant;
};

union ResultModelUnion {
    struct RocStr error;
    void* *model;
};

struct ResultModel {
    union ResultModelUnion payload;
    unsigned char disciminant;
};

// decodeModel
extern void roc__mainForHost_0_caller(const struct DecodeArg *arg, void* something, const struct ResultModel *resultModel);

// encodeModel
extern void roc__mainForHost_1_caller(void* *model, void* something, struct RocList *bytes);

// handleReadRequest
extern void roc__mainForHost_2_caller(const struct Request *request, void* *model,  void* something, const struct Response *response );

// handleWriteRequest
extern void roc__mainForHost_3_caller(const struct Request *request, void* *model,  void* something, const struct ResponseModel *responseModel );
