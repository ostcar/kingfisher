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
    struct RocStr value;
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
    struct RocList body;
    struct RocList headers;
    short unsigned int status;
};

// TODO: Can the union be directly inside ResultModel? Same above
union ResultModelUnion {
    struct RocStr error;
    void* *model;
};

struct ResultModel {
    union ResultModelUnion payload;
    unsigned char disciminant;
};

union ResultResponseUnion {
    struct Response response;
};

// TODO: Does this even have a disciminant??
struct ResultResponse {
    union ResultResponseUnion payload;
    unsigned char disciminant;
};

// TODO: Is this needed or can *model be directly used in MaybeModel?
union MaybeModelUnion {
    void* *model;
};

struct MaybeModel {
    union MaybeModelUnion payload;
    unsigned char disciminant;
};

struct ResultVoidVoid {
    unsigned char disciminant;
};

struct ResultVoidStr {
    struct RocStr payload;
    unsigned char disciminant;
};

// updateModel
extern void roc__mainForHost_2_caller(const struct RocList *events, const struct MaybeModel *maybeModel, void* something, const struct ResultModel *resultModel);

// respond
extern void roc__mainForHost_0_caller(const struct Request *request, void* *model,  void* something, void* captures );

size_t roc__mainForHost_0_result_size();
extern void roc__mainForHost_1_caller(char* flags, void* closure_data, struct ResultResponse *result);


