#include <stdio.h>
#include <emscripten.h>

struct FPDF_FILEACCESS {
  // File length, in bytes.
  unsigned long m_FileLen;

  // A function pointer for getting a block of data from a specific position.
  // Position is specified by byte offset from the beginning of the file.
  // The pointer to the buffer is never NULL and the size is never 0.
  // The position and size will never go out of range of the file length.
  // It may be possible for FPDFSDK to call this function multiple times for
  // the same position.
  // Return value: should be non-zero if successful, zero for error.
  int (*m_GetBlock)(void* param,
                    unsigned long position,
                    unsigned char* pBuf,
                    unsigned long size);

  // A custom pointer for all implementation specific data.  This pointer will
  // be used as the first parameter to the m_GetBlock callback.
  void* m_Param;
};

int main() {
  // Not doing anything here now.
  return 0;
}

EMSCRIPTEN_KEEPALIVE
int FPDF_LoadCustomDocument(struct FPDF_FILEACCESS* pFileAccess) {
  // Just loop through the bytes 1 by 1 and get them with m_GetBlock.

  int i;
  int read;
  unsigned char mybuffer[pFileAccess->m_FileLen];

  printf("Size to read %lu \n", pFileAccess->m_FileLen);

  for (i = 0; i < pFileAccess->m_FileLen; ++i)
  {
    printf("Reading byte number %d\n", i);

    // Pass a reference to the byte number we're at and always read 1 byte.
    pFileAccess->m_GetBlock(pFileAccess->m_Param, i, &mybuffer[i], 1);

    read++;
  }

  // Return the amount of read bytes.
  return read;
}
