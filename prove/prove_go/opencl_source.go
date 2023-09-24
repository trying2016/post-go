package post_go

var (
	openclSource = []byte(`

/*
This is an implementation of the AES algorithm, specifically CBC mode.
Block size can be chosen in aes.h - available choices are AES128.
The implementation is verified against the test vectors in:
  National Institute of Standards and Technology Special Publication 800-38A 2001 ED
https://github.com/kokke/tiny-AES-c
*/
/*****************************************************************************/
/* Includes:                                                                 */
/*****************************************************************************/
#define AES128_BLOCKLEN 16 // Block length in bytes - AES is 128b block only
#define AES128_KEYLEN 16   // Key length in bytes
#define AES128_KEYEXPSIZE 176
#define NONCES_PER_AES 16

typedef struct AES_ctx
{
    uint8_t RoundKey[AES128_KEYEXPSIZE];
    uint32_t nonce_group;
} AES_ctx;


struct Result {
    uint64_t index;
    uint64_t nonce;
};


/*****************************************************************************/
/* Defines:                                                                  */
/*****************************************************************************/
#define Nb 4  // The number of columns comprising a state in AES. This is a constant in AES. Value=4
#define Nk 4  // The number of 32 bit words in a key.
#define Nr 10 // The number of rounds in AES Cipher.

#define getSBoxValue(num) (sbox[(num)])

__constant uint8_t sbox[256] = {
    // 0     1    2      3     4    5     6     7      8    9     A      B    C     D     E     F
    0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe, 0xd7, 0xab, 0x76,
    0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4, 0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0,
    0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7, 0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15,
    0x04, 0xc7, 0x23, 0xc3, 0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75,
    0x09, 0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3, 0x2f, 0x84,
    0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe, 0x39, 0x4a, 0x4c, 0x58, 0xcf,
    0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85, 0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8,
    0x51, 0xa3, 0x40, 0x8f, 0x92, 0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2,
    0xcd, 0x0c, 0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19, 0x73,
    0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14, 0xde, 0x5e, 0x0b, 0xdb,
    0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2, 0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79,
    0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5, 0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08,
    0xba, 0x78, 0x25, 0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
    0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86, 0xc1, 0x1d, 0x9e,
    0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e, 0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf,
    0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42, 0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16};

__constant uint8_t Rcon[11] = {
    0x8d, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x1b, 0x36};

static inline void FirstAddRoundKey(uint8_t round, const __global uint8_t *state,  uint8_t *out_state, const __global uint8_t *RoundKey)
{
    for (uint8_t i = 0; i < 4; ++i)
    {
        for (uint8_t j = 0; j < 4; ++j)
        {
            out_state[i * Nb + j] = state[i * Nb + j] ^ RoundKey[round * Nb * 4 + i * Nb + j];
        }
    }
}

static inline void AddRoundKey(uint8_t round, uint8_t *state, const __global uint8_t *RoundKey)
{
    for (uint8_t i = 0; i < 4; ++i)
    {
        for (uint8_t j = 0; j < 4; ++j)
        {
            state[i * Nb + j] ^= RoundKey[round * Nb * 4 + i * Nb + j];
        }
    }
}

static inline void SubBytes(uint8_t *state)
{
    for (uint8_t i = 0; i < 4; ++i)
    {
        for (uint8_t j = 0; j < 4; ++j)
        {
            state[j * Nb + i] = getSBoxValue(state[j * Nb + i]);
        }
    }
}

static inline void ShiftRows(uint8_t *state)
{
    uint8_t temp;

    // Rotate first row 1 columns to left
    temp = state[0*4+1];
    state[0*4+1] = state[1*4+1];
    state[1*4+1] = state[2*4+1];
    state[2*4+1] = state[3*4+1];
    state[3*4+1] = temp;

    // Rotate second row 2 columns to left
    temp = state[0*4+2];
    state[0*4+2] = state[2*4+2];
    state[2*4+2] = temp;

    temp = state[1*4+2];
    state[1*4+2] = state[3*4+2];
    state[3*4+2] = temp;

    // Rotate third row 3 columns to left
    temp = state[0*4+3];
    state[0*4+3] = state[3*4+3];
    state[3*4+3] = state[2*4+3];
    state[2*4+3] = state[1*4+3];
    state[1*4+3] = temp;
}

static inline uint8_t xtime(uint8_t x)
{
    return ((x << 1) ^ (((x >> 7) & 1) * 0x1b));
}

static inline void MixColumns(uint8_t *state)
{
    for (uint8_t i = 0; i < 4; ++i)
    {
        uint8_t t = state[i * Nb + 0];
        uint8_t Tmp = state[i * Nb + 0] ^ state[i * Nb + 1] ^ state[i * Nb + 2] ^ state[i * Nb + 3];
        uint8_t Tm = state[i * Nb + 0] ^ state[i * Nb + 1];
        Tm = xtime(Tm);
        state[i * Nb + 0] ^= Tm ^ Tmp;
        Tm = state[i * Nb + 1] ^ state[i * Nb + 2];
        Tm = xtime(Tm);
        state[i * Nb + 1] ^= Tm ^ Tmp;
        Tm = state[i * Nb + 2] ^ state[i * Nb + 3];
        Tm = xtime(Tm);
        state[i * Nb + 2] ^= Tm ^ Tmp;
        Tm = state[i * Nb + 3] ^ t;
        Tm = xtime(Tm);
        state[i * Nb + 3] ^= Tm ^ Tmp;
    }
}

#define Multiply(x, y)                         \
    (((y & 1) * x) ^                           \
     ((y >> 1 & 1) * xtime(x)) ^               \
     ((y >> 2 & 1) * xtime(xtime(x))) ^        \
     ((y >> 3 & 1) * xtime(xtime(xtime(x)))) ^ \
     ((y >> 4 & 1) * xtime(xtime(xtime(xtime(x))))))

#define getSBoxInvert(num) (rsbox[(num)])

static inline void Cipher(const __global uint8_t *state, uint8_t *out_state, const __global uint8_t *RoundKey)
{
    uint8_t round = 1;
    FirstAddRoundKey(0, state, out_state, RoundKey);
    // for(int i=0;i<16;i++){
    //     printf("OpenCl FirstAddRoundKey[%d]=%d\n", i, out_state[i]);
    // }
    for (;; ++round)
    {
        SubBytes(out_state);
        // for(int i=0;i<16;i++){
        //     printf("OpenCl SubBytes[%d]=%d\n", i, out_state[i]);
        // }
        ShiftRows(out_state);
        // for(int i=0;i<16;i++){
        //     printf("OpenCl SubBytes[%d]=%d\n", i, out_state[i]);
        // }
        if (round == Nr)
            break;
        MixColumns(out_state);
        AddRoundKey(round, out_state, RoundKey);
    }
    AddRoundKey(Nr, out_state, RoundKey);
}

void AES_CBC_encrypt_buffer(__global uint8_t* roundKey, const __global uint8_t *in, uint8_t *out, size_t length)
{
    Cipher(in, out, roundKey);
}

void AES_init_ctx(AES_ctx *ctx, const uint8_t *key)
{
    uint8_t tempa[4];
    for (int i = 0; i < Nk; ++i)
    {
        ctx->RoundKey[i * 4 + 0] = key[i * 4 + 0];
        ctx->RoundKey[i * 4 + 1] = key[i * 4 + 1];
        ctx->RoundKey[i * 4 + 2] = key[i * 4 + 2];
        ctx->RoundKey[i * 4 + 3] = key[i * 4 + 3];
    }
    for (int i = Nk; i < Nb * (Nr + 1); ++i)
    {
        {
            int k = (i - 1) * 4;
            tempa[0] = ctx->RoundKey[k + 0];
            tempa[1] = ctx->RoundKey[k + 1];
            tempa[2] = ctx->RoundKey[k + 2];
            tempa[3] = ctx->RoundKey[k + 3];
        }
        if (i % Nk == 0)
        {
            {
                const uint8_t u8tmp = tempa[0];
                tempa[0] = tempa[1];
                tempa[1] = tempa[2];
                tempa[2] = tempa[3];
                tempa[3] = u8tmp;
            }
            {
                tempa[0] = getSBoxValue(tempa[0]);
                tempa[1] = getSBoxValue(tempa[1]);
                tempa[2] = getSBoxValue(tempa[2]);
                tempa[3] = getSBoxValue(tempa[3]);
            }
            tempa[0] = tempa[0] ^ Rcon[i / Nk];
        }
        int j = i * 4;
        int k = (i - Nk) * 4;
        ctx->RoundKey[j + 0] = ctx->RoundKey[k + 0] ^ tempa[0];
        ctx->RoundKey[j + 1] = ctx->RoundKey[k + 1] ^ tempa[1];
        ctx->RoundKey[j + 2] = ctx->RoundKey[k + 2] ^ tempa[2];
        ctx->RoundKey[j + 3] = ctx->RoundKey[k + 3] ^ tempa[3];
    }
}

// Calculate nonce value given nonce group and its offset within the group.
uint32_t calc_nonce(uint32_t nonce_group, uint32_t per_aes, uint32_t offset) {
    return nonce_group * per_aes + offset % per_aes;
}

// LSB part of the difficulty is checked with second sequence of AES ciphers.
void check_lsb(__global struct AES_ctx* nonce_cipher, uint64_t difficulty_lsb, const __global uint8_t *label, uint32_t nonce, uint64_t nonce_offset, uint64_t base_index, __global struct Result* out_index, __global int *out_nonce_offset) {
    uint64_t temp[2];
    __global struct AES_ctx *lazy = nonce_cipher;
    AES_CBC_encrypt_buffer(lazy->RoundKey, label, (uint8_t *)temp, AES128_BLOCKLEN);
    uint64_t lsb = temp[0] & 0x00ffffffffffffff;
    if (lsb < difficulty_lsb) {
        uint64_t index = base_index;// + (nonce_offset / NONCES_PER_AES);
        int nonce_index = atomic_add(out_nonce_offset, 1);
        
        __global struct Result* out = out_index + nonce_index;
        out->index = index;
        out->nonce = nonce;
    }
}

__kernel
void prove_part(__global struct AES_ctx* d_group_cipher, 
           __global struct AES_ctx* d_nonce_cipher, 
           uint64_t base_index, 
           uint32_t group_counter,
           uint32_t difficulty_msb,
           uint64_t difficulty_lsb,
           const __global uint8_t *in, 
           __global struct Result* out_index, 
           __global int *out_nonce_offset, 
           uint32_t total) {
    
    size_t g_index = get_global_id(0);
    size_t index = g_index / group_counter;
    if (g_index >= total)
        return;
    // if(index == 0){
    //     // 打印data前30个字符
    //     // printf("opencl data: [");
    //     // for(int i=0;i<30;i++){
    //     //     printf("%d, ", in[i]);
    //     // }
    //     // printf("]\n");

    //     printf("index: %d group_counter: %d difficulty_msb: %d difficulty_lsb: %lu total: %lu\n",
    //         index, group_counter, difficulty_msb, difficulty_lsb, total);
    // }

    uint8_t temp[AES128_BLOCKLEN];
    const __global uint8_t* chunk = in + index * AES128_BLOCKLEN;

    //printf("index 1: %d\n", index);
    //printf("opencl sizeof(struct AES_ctx) = %d\n", sizeof( AES_ctx));

    int group = g_index % group_counter;
    __global struct AES_ctx *cipher = d_group_cipher + group;
    AES_CBC_encrypt_buffer(cipher->RoundKey, chunk, temp, AES128_BLOCKLEN);

    for (int offset = 0; offset < AES128_BLOCKLEN; offset++) {
        uint8_t msb = temp[offset];
        if (msb <= difficulty_msb) {
            // 打印msb, nonce_group, offset, difficulty_msb, difficulty_lsb
            // if(index == 0){
            //     printf("msb %d nonce_group: %d offset: %d difficulty_msb: %d difficulty_lsb: %d\n", msb, cipher->nonce_group, offset, difficulty_msb, difficulty_lsb);
            // }
            // printf("msb %d difficulty_msb: %d\n", msb, difficulty_msb);
            if (msb == difficulty_msb) {
                uint32_t nonce = calc_nonce(cipher->nonce_group, NONCES_PER_AES, offset);
                //printf("msb %d nonce_group: %d offset: %d difficulty_msb: %d difficulty_lsb: %d\n", msb, cipher->nonce_group, offset, difficulty_msb, difficulty_lsb);
                check_lsb(d_nonce_cipher + group * AES128_BLOCKLEN + offset, difficulty_lsb, chunk, nonce, offset, base_index + index, out_index, out_nonce_offset);
            } else {
                uint32_t nonce = calc_nonce(cipher->nonce_group, NONCES_PER_AES, offset);
                int nonce_index = atomic_add(out_nonce_offset, 1);
                    __global struct Result* out = out_index + nonce_index;
                out->index = base_index + index;
                out->nonce = nonce;
                // printf index, nonce
                // printf("Nonce [%d] = %d\n", index, nonce);
            }
        }
    }
}


__kernel
void prove(__global struct AES_ctx* d_group_cipher, 
           __global struct AES_ctx* d_nonce_cipher, 
           uint64_t base_index, 
           uint32_t group_counter,
           uint32_t difficulty_msb,
           uint64_t difficulty_lsb,
           const __global uint8_t *in, 
           __global struct Result* out_index, 
           __global int *out_nonce_offset, 
           uint32_t total) {
    
    size_t index = get_global_id(0);
    if (index >= total)
        return;
    // if(index == 0){
    //     // 打印data前30个字符
    //     // printf("opencl data: [");
    //     // for(int i=0;i<30;i++){
    //     //     printf("%d, ", in[i]);
    //     // }
    //     // printf("]\n");

    //     printf("index: %d group_counter: %d difficulty_msb: %d difficulty_lsb: %lu total: %lu\n",
    //         index, group_counter, difficulty_msb, difficulty_lsb, total);
    // }

    uint8_t temp[AES128_BLOCKLEN];
    const __global uint8_t* chunk = in + index * AES128_BLOCKLEN;

    //printf("index 1: %d\n", index);
    //printf("opencl sizeof(struct AES_ctx) = %d\n", sizeof( AES_ctx));

    for (int i = 0; i < group_counter; i++) {
        // for(int j=0;j<16;j++) {
        //     printf("Opencl In:[%d]=%d\n", j, chunk[j]);
        // }
        __global struct AES_ctx *cipher = d_group_cipher + i;
        AES_CBC_encrypt_buffer(cipher->RoundKey, chunk, temp, AES128_BLOCKLEN);
        // for(int j=0;j<AES128_KEYEXPSIZE;j++){
        //     printf("Opencl RoundKey[%d]=%d\n", j, cipher->RoundKey[j]);
        // }
        // 打印chunk[16],temp[16]
        // if(index == 0){
        //     printf("chunk [%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d]", chunk[0], chunk[1], chunk[2], chunk[3], chunk[4], chunk[5], chunk[6], chunk[7], chunk[8], chunk[9], chunk[10], chunk[11]);
        //     printf("temp [%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d]\n", temp[0], temp[1], temp[2], temp[3], temp[4], temp[5], temp[6], temp[7], temp[8], temp[9], temp[10], temp[10]);
        // }

        // if(index == 0){
        //     printf("Opencl nonce_group: %d sizeof(AES_ctx)=%d\n", cipher->nonce_group, sizeof(AES_ctx));
        // }

        for (int offset = 0; offset < AES128_BLOCKLEN; offset++) {
            uint8_t msb = temp[offset];
            if (msb <= difficulty_msb) {
                // 打印msb, nonce_group, offset, difficulty_msb, difficulty_lsb
                // if(index == 0){
                //     printf("msb %d nonce_group: %d offset: %d difficulty_msb: %d difficulty_lsb: %d\n", msb, cipher->nonce_group, offset, difficulty_msb, difficulty_lsb);
                // }
                // printf("msb %d difficulty_msb: %d\n", msb, difficulty_msb);
                if (msb == difficulty_msb) {
                    uint32_t nonce = calc_nonce(cipher->nonce_group, NONCES_PER_AES, offset);
                    //printf("msb %d nonce_group: %d offset: %d difficulty_msb: %d difficulty_lsb: %d\n", msb, cipher->nonce_group, offset, difficulty_msb, difficulty_lsb);
                    check_lsb(d_nonce_cipher + i * AES128_BLOCKLEN + offset, difficulty_lsb, chunk, nonce, offset, base_index + index, out_index, out_nonce_offset);
                } else {
                    uint32_t nonce = calc_nonce(cipher->nonce_group, NONCES_PER_AES, offset);
                    int nonce_index = atomic_add(out_nonce_offset, 1);
                     __global struct Result* out = out_index + nonce_index;
                    out->index = base_index + index;
                    out->nonce = nonce;
                    // printf index, nonce
                    // printf("Nonce [%d] = %d\n", index, nonce);
                }
            }
        }
    }
}

`)
)
