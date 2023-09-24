#ifndef POST_CAPI_H
#define POST_CAPI_H

#include <stdint.h>


#ifdef __cplusplus
extern "C" {
#endif

#define MAX_RESULT_SIZE 1024*1024
#define LOG_INFO 1
#define LOG_ERROR 2

// 按照4字节对齐
#pragma pack(push, 4)
typedef struct {
    uint64_t index;
    uint32_t nonce;
} Result;
#pragma pack(pop)

typedef struct post_gpu  post_gpu;

typedef void(*log_callback)(int level, const char* message);

void set_log_callback(log_callback callback);

post_gpu* post_create(int device,
                      int start,
                      int nonces,
                      uint8_t *ciphers_keys,
                      uint8_t *lazy_ciphers_keys,
                      uint64_t difficulty_lsb,
                      uint8_t difficulty_msb,
                      int input_size,
                      const char *sources,
                      int source_size);

void post_destroy(post_gpu* ctx);

int post_prove(post_gpu* ctx, uint64_t base_index, uint8_t* data, int data_size);

void post_get_results(post_gpu* ctx, int index, Result* result);

int post_device_count();
int post_device_name(int device, char* name, int size);

#ifdef __cplusplus
}
#endif

#endif /* POST_CAPI_H */
