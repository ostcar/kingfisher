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

struct Program {
    void* init;
};

struct BodyMimeType {
    struct RocStr body;
    struct RocStr mimeType;
};

union RequestBodyUnion {
    struct BodyMimeType body;
};

struct RequestBody {
    union RequestBodyUnion payload;
    unsigned char discriminant;
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
    struct RocStr value;
};

struct Request {
    struct RequestBody body;
    struct RocList headers;
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

extern void roc__mainForHost_1_exposed_generic(const struct Program *program);

// handleReadRequest
extern void roc__mainForHost_0_caller(const struct Request *request, void* *model,  void* something, const struct Response *response );

// handleWriteRequest
extern void roc__mainForHost_1_caller(const struct Request *request, void* *model,  void* something, const struct ResponseModel *responseModel );
