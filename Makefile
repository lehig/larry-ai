.PHONY: up down test logs

up:
	powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 up
	@echo.
	@echo Performing smoke tests...
	powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 test

down:
	powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 down

test:
	powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 test

logs:
	powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 logs
