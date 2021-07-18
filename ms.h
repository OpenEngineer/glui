#include <stdio.h>
#include <dwmapi.h>
#include <SDL2/SDL_events.h>
#include <SDL2/SDL_syswm.h>

void SetWindowAttributes(HWND hwnd) {
   BOOL fForceIconic = TRUE;
   BOOL fHasIconicBitmap = TRUE;

   DwmSetWindowAttribute(
       hwnd,
       DWMWA_FORCE_ICONIC_REPRESENTATION,
       &fForceIconic,
       sizeof(fForceIconic));

   DwmSetWindowAttribute(
       hwnd,
       DWMWA_HAS_ICONIC_BITMAP,
       &fHasIconicBitmap,
       sizeof(fHasIconicBitmap));

  printf("succesfully called DwmSetWindowAttribute\n");
}

UINT GetSysWMmsgType(SDL_SysWMmsg* msg) {
  return msg->msg.win.msg;
}

uint32_t GetIconicThumbnailMaxWidth(SDL_SysWMmsg* msg) {
  return HIWORD(msg->msg.win.lParam);
}

uint32_t GetIconicThumbnailMaxHeight(SDL_SysWMmsg* msg) {
  return LOWORD(msg->msg.win.lParam);
}

void SetIconicThumbnail(HWND hwnd, uint32_t w, uint32_t h, const BYTE* data) {
  // bitmap creation based on windows7 sdk sample
  HDC hdcMem = CreateCompatibleDC(NULL);
  if (hdcMem == NULL) {
    fprintf(stderr, "something went wrong when preparing thumbnail bitmap\n");
    return;
  }

  BITMAPINFO bmi;
  ZeroMemory(&bmi.bmiHeader, sizeof(BITMAPINFOHEADER));
  bmi.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
  bmi.bmiHeader.biWidth = w;
  bmi.bmiHeader.biHeight = -h;
  bmi.bmiHeader.biPlanes = 1;
  bmi.bmiHeader.biBitCount = 32;

  PBYTE pbDS = NULL;
  HBITMAP bmp = CreateDIBSection(hdcMem, &bmi, DIB_RGB_COLORS, (VOID**)&pbDS, NULL, 0);
  if (bmp == NULL) {
    fprintf(stderr, "something went wrong when allocating thumbnail bitmap\n");
    return;
  }

  for (int y = 0; y < (int)h; y++) {
    for (int x = 0; x < (int)w; x++) {
      int k = x*(int)h + y;
      pbDS[0] = data[k*4+0];
      pbDS[1] = data[k*4+1];
      pbDS[2] = data[k*4+2];
      pbDS[3] = data[k*4+3];
      pbDS += 4;
    }
  }

  DeleteDC(hdcMem);

  //HBITMAP bmp = CreateBitmap(w, h, 1, 32, data);

  HRESULT res = DwmSetIconicThumbnail(hwnd, bmp, 0);
  if (res != S_OK) {
    fprintf(stderr, "something went wrong when setting iconic thumbnail (%d, %#06x, %#010x)\n", res, res, bmp);
  }

  DeleteObject(bmp);
}
