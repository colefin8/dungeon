// go to https://convertico.com/image-to-ascii/
// change Width to 64px
// check Invert colors
// change Color mode to Full Color
// change Character set to Detailed
// inspect a character in the output window
// right-click the outer <pre> and select Copy OuterHTML
// paste into a text file
// format like frame1.txt or frame2.txt
// run "py rep.py frameX.txt frameX.bin"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define FRAME_CAP 1040 * 26
#define SLEEP_TIME 500000
// #define SLEEP_TIME 150000 // pretty fast
// #define SLEEP_TIME 60000 // fast

void printImg();
void printImg2();

int main(int argc, char** argv) {
  // frame 1
  FILE *f = fopen("frame1.bin", "rb");
  // FILE *f = fopen("frame1-mono.txt", "rb");
  if (!f) return 1;
  fseek(f, 0, SEEK_END);
  long size1 = ftell(f);
  rewind(f);
  if (size1 <= 0) return 2;
  char *buffer1 = malloc(size1);
  if (!buffer1) return 3;
  if (fread(buffer1, 1, size1, f) != (size_t)size1)
      return 4;
  fclose(f);

  // frame 2
  f = fopen("frame2.bin", "rb");
  // f = fopen("frame2-mono.txt", "rb");
  if (!f) return 5;
  fseek(f, 0, SEEK_END);
  long size2 = ftell(f);
  rewind(f);
  if (size2 <= 0) return 6;
  char *buffer2 = malloc(size2);
  if (!buffer2) return 7;
  if (fread(buffer2, 1, size2, f) != (size_t)size2)
      return 8;
  fclose(f);

  // frame 3
  // f = fopen("frame3.bin", "rb");
  f = fopen("frame3-mono.txt", "rb");
  if (!f) return 5;
  fseek(f, 0, SEEK_END);
  long size3 = ftell(f);
  rewind(f);
  if (size3 <= 0) return 6;
  char *buffer3 = malloc(size3);
  if (!buffer3) return 7;
  if (fread(buffer3, 1, size3, f) != (size_t)size3)
      return 8;
  fclose(f);

  int frameNum = 0;
  write(1, "\x1b[48;2;8;8;8m", 13);
  while (1) {
    write(1, "\x1b[H\x1b[J", 6);

    // int frame = frameNum % 3;
    int frame = frameNum % 2;
    switch (frame) {
      case 0:
        write(1, buffer1, size1);
        break;
      case 1:
        write(1, buffer2, size2);
        break;
      // case 2:
      //   write(1, buffer3, size3);
      //   break;
    }
    printf("\x1b[37mframeNum: %i\n", frameNum);
    printf("frame: %i\n", frame);

    frameNum++;
    usleep(SLEEP_TIME);
  }

  return 0;
}
