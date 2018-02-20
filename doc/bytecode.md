# Schego bytecode format documentation

# Basic structure

Similar to the CLR and JVM, each individual bytecode instruction is one byte long. Each byte is either a single opcode,
or a part of a literal. This includes multi-byte literals such as ints, strings, and so forth, which are simply stored in
contiguous memory locations.
There are no registers to speak of; Schego's VM is entirely stack-based. The standard push and pop instructions
are supported; writing to the heap or to the local frame is also possible.

## Opcode structure
All opcodes are of a fixed size (1 byte), with potential arguments for that opcode following immediately after.
The number of arguments each opcode can have is always fixed.

## Basic supported datatypes
* Boolean (1 byte)
* UTF-8 character (1 to 4 bytes)
* 64-bit signed integer (8 bytes, little endian)
* 64-bit signed double precision float (8 bytes, little endian)
* UTF-8 null-terminated string (8 bytes per character, variable size)
* List (8 bytes to indicate size plus an additional 8 bytes per element to indicate location in memory)


# Opcode reference
(b)ool (c)har (i)nteger (d)ouble (s)tring (l)ist

## pushb
Opcode: **0x01**

Pushes a Boolean literal onto the stack.
The literal is the byte immediately following the opcode.

## pushc
Opcode: **0x02**

Pushes a UTF-8 character literal onto the stack.
The literal is the byte immediately following the opcode.

## pushi
Opcode: **0x03**

Pushes a 64-bit integer literal onto the stack.
The literal is the next 8 bytes immediately following the opcode.

## pushd
Opcode: **0x04**

Pushes a 64-bit double precision float onto the stack.
The literal is the next 8 bytes immediately following the opcode.

## pushs
Opcode: **0x05**

Pushes a null-terminated UTF-8 string onto the stack.
Every 8 bytes up until a null terminator is found is considered part of the string.
*Warning:* The behavior of this opcode if **pushs** does not find a null terminator is undefined!

## cons
Opcode: **0x06**

Pushes an empty list cell onto the stack.

## dup
Opcode: **0x07**
Duplicates whatever literal is on top of the stack.

## hstoreb
Opcode: **0x08**

Stores a boolean from the top of the stack in the heap.
The next 8 bytes immediately following this instruction represent the reference in heap memory
that the value should be stored in.

## hstorec
Opcode: **0x09**

Stores a UTF-8 character from the top of the stack in the heap.
The next 8 bytes immediately following this instruction represent the reference in heap memory
that the value shold be stored in.

## hstorei
Opcode: **0x0A**

Stores a 64-bit integer from the top of the stack in the heap.
The next 8 bytes immediately following this instruction represent the reference in heap memory
that the value shold be stored in.

## hstored
Opcode: **0x0B**

Stores a 64-bit double precision float from the top of the stack in the heap.
The next 8 bytes immediately following this instruction represent the reference in heap memory
that the value shold be stored in.

## hstores
Opcode: **0x0C**

Stores a null-terminated UTF-8 string from the top of the stack in the heap.
The next 8 bytes immediately following this instruction represent the reference in heap memory
that the value shold be stored in.

## hstorel
Opcode: **0x0D**

Stores a list from the top of the stack in the heap.
The next 8 bytes immediately following this instruction represent the reference in heap memory
that the value shold be stored in.

## lstoreb
Opcode: **0x0E**

Stores a boolean from the top of the stack in the local frame.
The next 4 bytes immediately following this instruction represent the reference in local memory
that the value shold be stored in.
## lstorec
Opcode: **0x0F**

Stores a UTF-8 character literal from the top of the stack in the local frame.
The next 4 bytes immediately following this instruction represent the reference in local memory
that the value shold be stored in. 
## lstorei
Opcode: **0x10**

Stores a 64-bit integer from the top of the stack in the local frame.
The next 4 bytes immediately following this instruction represent the reference in local memory
that the value shold be stored in.

## lstored
Opcode: **0x11**

Stores a 64-bit double precision float from the top of the stack in the local frame.
The next 4 bytes immediately following this instruction represent the reference in local memory
that the value shold be stored in.

## lstores
Opcode: **0x12**

Stores a UTF-8 null-terminated string from the top of the stack in the local frame.
The next 4 bytes immediately following this instruction represent the reference in local memory
that the value shold be stored in.

## lstorel
Opcode: **0x13**

Stores a list from the top of the stack into the local frame.
The next 4 bytes immediately following this instruction represent the reference in local memory
that the value shold be stored in.

## hloadb
Opcode: **0x14**
## hloadc
Opcode: **0x15**
## hloadi
Opcode: **0x16**
## hloadd
Opcode: **0x17**
## hloads
Opcode: **0x18**
## hloadl
Opcode: **0x19**
## lloadb
Opcode: **0x1A**
## lloadc
Opcode: **0x1B**
## lloadi
Opcode: **0x1C**
## lloadd
Opcode: **0x1D**
## lloads
Opcode: **0x1E**
## lloadl
Opcode: **0x1F**
## hnewb
Opcode: **0x20**
## hnewc
Opcode: **0x21**
## hnewi
Opcode: **0x22**
## hnewd
Opcode: **0x23**
## hnews
Opcode: **0x24**
## hnewl
Opcode: **0x25**
## lnewb
Opcode: **0x26**
## lnewc
Opcode: **0x27**
## lnewi
Opcode: **0x28**
## lnewd
Opcode: **0x29**
## lnews
Opcode: **0x2A**
## lnewl
Opcode: **0x2B**

## jmp
Opcode: **0x2C**
## jne
Opcode: **0x2D**
## jeq
Opcode: **0x2E**
## jlt
Opcode: **0x2F**
## jlte
Opcode: **0x30**
## jgt
Opcode: **0x31**
## jgte
Opcode: **0x32**
## jal
Opcode: **0x33**
## jr
Opcode: **0x34**

## addc
Opcode: **0x35**
## addi
Opcode: **0x36**
## addd
Opcode: **0x37**
## adds
Opcode: **0x36**

## subc
Opcode: **0x37**
## subi
Opcode: **0x38**
# subd
Opcode: **0x39**

## muli
Opcode: **0x3A**
## muld
Opcode: **0x3B**
## divc
Opcode: **0x3C**
## divi
Opcode: **0x3D**
## divd
Opcode: **0x3E**

## cmpc
Opcode: **0x3F**
## cmpi
Opcode: **0x40**
## cmpd
Opcode: **0x41**
## cmps
Opcode: **0x42**

## syscall
Opcode: **0x43**

Performs a system call into the interpreter, with the first byte immediately following
the opcode indicating which syscall to perform. Each syscall has its own set of arguments,
which are taken from the stack. The set of valid syscalls are:

* **0x01** Print boolean from the stack to standard output.
* **0x02** Print character from the stack to standard output.
* **0x03** Print integer from the stack to standard output.
* **0x04** Print double from the stack to standard output.
* **0x05** Print string from the stack to standard output.
* **0x06** Exits the interpreter. The top integer on the stack is used for the return code.

## hsmnem
Opcode: **0x44**

Sets the 2-byte mnemonic reference immediately following the opcode to the address of the
second 2-byte mnemonic reference.

## lsmnem
Opcode: **0x45**

## cmpl
Opcode: **0x46**

Compares the data addresses of two list cells.

## hcar
Opcode: **0x47**

## lcar
Opcode: **0x48**

## hcdr
Opcode: **0x49**

## lcdr
Opcode: **0x4A**

## hscar
Opcode: **0x4B**

## lscar
Opcode: **0x4C**

## hscdr
Opcode: **0x4D**

## lscdr
Opcode: **0x4E**
