#pragma once
/*++
This code is subjected to Copyright.
Owner of this code is sakshyam@seknox.com
Seknox Cybersecurity Pvt. Ltd.
--*/




#include <stdio.h>
#include <vector>
#include <string.h>
#include <winhttp.h>
#include <strsafe.h>
#include <iostream>
#include <windows.h>
#include <tchar.h>

#define INFO_BUFFER_SIZE 32767



BOOL WriteLogFile(LPWSTR Stringval);
std::string sendRequest(std::string user, std::string totp);
//BOOL createFileFunc(LPSTR stringval);
//std::string sendRequest(std::string user);