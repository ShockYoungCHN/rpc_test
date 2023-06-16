#include "textflag.h"

DATA	msg+0(SB)/13, $"Hello, world!"
DATA	newline+0(SB)/1, $0x0a

GLOBL msg(SB), RODATA, $13
GLOBL newline(SB), RODATA, $1


TEXT	·out(SB), NOSPLIT, $0-0
	LEAQ	msg(SB), AX      // 将消息的地址保存在BX寄存器中
	MOVL	$13, BX          // 消息的长度保存在CX寄存器中
	CALL    runtime·printstring(SB)
	RET
