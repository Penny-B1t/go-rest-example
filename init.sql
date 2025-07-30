-- IoT Device 데이터베이스 초기화 스크립트

-- 데이터베이스 사용
USE iot_device_db;

-- devices 테이블 생성
CREATE TABLE IF NOT EXISTS devices (
    InternalID BIGINT AUTO_INCREMENT PRIMARY KEY,
    ProductNumber VARCHAR(50) NOT NULL UNIQUE,
    MacAddress VARCHAR(17) NOT NULL UNIQUE,
    FirmwareVersion VARCHAR(20) NOT NULL,
    LastSeenAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ReTry INT DEFAULT 0,
    UpdateCheck INT DEFAULT 0,
    Status ENUM('Ready', 'PowerOn', 'PowerOff', 'ERROR') DEFAULT 'Ready',
    INDEX idx_product_number (ProductNumber),
    INDEX idx_mac_address (MacAddress),
    INDEX idx_status (Status),
    INDEX idx_last_seen (LastSeenAt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- reports 테이블 생성
CREATE TABLE IF NOT EXISTS reports (
    ReportID BIGINT AUTO_INCREMENT PRIMARY KEY,
    ProductNumber VARCHAR(50) NOT NULL,
    BatteryPercent INT NOT NULL CHECK (BatteryPercent >= 0 AND BatteryPercent <= 100),
    Lat DECIMAL(10, 8) NOT NULL,
    Lon DECIMAL(11, 8) NOT NULL,
    TemperatureCelsius DECIMAL(5, 2) NOT NULL,
    IP VARCHAR(45) NOT NULL,
    ErrorCode INT DEFAULT 0,
    ReportAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ReportedStatus ENUM('Ready', 'PowerOn', 'PowerOff', 'ERROR') NOT NULL,
    INDEX idx_product_number (ProductNumber),
    INDEX idx_report_at (ReportAt),
    INDEX idx_status (ReportedStatus),
    INDEX idx_error_code (ErrorCode),
    FOREIGN KEY (ProductNumber) REFERENCES devices(ProductNumber) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 샘플 데이터 삽입 (테스트용)
INSERT INTO devices (ProductNumber, MacAddress, FirmwareVersion, Status) VALUES
('DEV001', '00:1B:44:11:3A:B7', 'v1.0.0', 'Ready'),
('DEV002', '00:1B:44:11:3A:B8', 'v1.1.0', 'Ready'),
('DEV003', '00:1B:44:11:3A:B9', 'v1.0.5', 'PowerOn');

-- 권한 설정 (선택사항 - 보안 강화를 위해)
-- CREATE USER IF NOT EXISTS 'iot_app'@'%' IDENTIFIED BY 'secure_app_password';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON iot_device_db.* TO 'iot_app'@'%';
-- FLUSH PRIVILEGES;