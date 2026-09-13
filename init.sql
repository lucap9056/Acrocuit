CREATE TABLE SpaceGroup (
    Id INT IDENTITY(1, 1) PRIMARY KEY,
    Name NVARCHAR(255) NOT NULL
);

CREATE TABLE Space (
    Id INT IDENTITY(1, 1) PRIMARY KEY,
    Name NVARCHAR(255) NOT NULL,
    SpaceGroupId INT NOT NULL,
    DisplayOrder INT NOT NULL,
    BackgroundImageUpdatedAt DATETIME2 NOT NULL DEFAULT GETUTCDATE(),
    CONSTRAINT FK_Space_SpaceGroup FOREIGN KEY (SpaceGroupId) REFERENCES SpaceGroup(Id),
    CONSTRAINT UQ_SpaceGroup_Order UNIQUE (SpaceGroupId, DisplayOrder),
    CONSTRAINT UQ_Space_Id_SpaceGroupId UNIQUE (Id, SpaceGroupId)
);

CREATE INDEX IX_Space_SpaceGroupId ON Space(SpaceGroupId);

CREATE TABLE BreakerGroup (
    Id INT IDENTITY(1, 1) PRIMARY KEY,
    Name NVARCHAR(255) NOT NULL,
    SpaceId INT NOT NULL,
    SpaceGroupId INT NOT NULL,
    CONSTRAINT FK_BreakerGroup_Space FOREIGN KEY (SpaceId, SpaceGroupId) REFERENCES Space(Id, SpaceGroupId),
    CONSTRAINT UQ_BreakerGroup_Id_SpaceGroupId UNIQUE (Id, SpaceGroupId)
);

CREATE INDEX IX_BreakerGroup_SpaceId ON BreakerGroup(SpaceId);

CREATE TABLE BreakerGroupPos (
    BreakerGroupId INT PRIMARY KEY,
    PositionX INT NOT NULL,
    PositionY INT NOT NULL,
    PositionZ INT NOT NULL,
    FOREIGN KEY (BreakerGroupId) REFERENCES BreakerGroup(Id) ON DELETE CASCADE
);

CREATE TABLE Breaker (
    Id INT IDENTITY(1, 1) PRIMARY KEY,
    Name NVARCHAR(255) NOT NULL,
    BreakerGroupId INT NOT NULL,
    SpaceGroupId INT NOT NULL,
    DisplayOrder INT NOT NULL,
    UpstreamBreaker INT,
    CONSTRAINT FK_Breaker_BreakerGroup FOREIGN KEY (BreakerGroupId, SpaceGroupId) REFERENCES BreakerGroup(Id, SpaceGroupId),
    CONSTRAINT UQ_BreakerGroup_Order UNIQUE (BreakerGroupId, DisplayOrder),
    CONSTRAINT UQ_Breaker_Id_BreakerGroupId UNIQUE (Id, BreakerGroupId),
    CONSTRAINT FK_Breaker_UpstreamBreaker FOREIGN KEY (UpstreamBreaker) REFERENCES Breaker(Id),
    CONSTRAINT CHK_Breaker_NotSelfUpstream CHECK (
        UpstreamBreaker IS NULL
        OR UpstreamBreaker <> Id
    )
);

CREATE INDEX IX_Breaker_BreakerGroupId ON Breaker(BreakerGroupId);

CREATE INDEX IX_Breaker_UpstreamBreaker ON Breaker(UpstreamBreaker);

CREATE TABLE Device (
    Id INT IDENTITY(1, 1) PRIMARY KEY,
    Name NVARCHAR(255) NOT NULL,
    SpaceId INT NOT NULL,
    CONSTRAINT FK_Device_Space FOREIGN KEY (SpaceId) REFERENCES Space(Id) ON DELETE CASCADE
);

CREATE TABLE DevicePos (
    DeviceId INT PRIMARY KEY,
    PositionX INT NOT NULL,
    PositionY INT NOT NULL,
    PositionZ INT NOT NULL,
    FOREIGN KEY (DeviceId) REFERENCES Device(Id) ON DELETE CASCADE
);

CREATE INDEX IX_Device_SpaceId ON Device(SpaceId);

CREATE TABLE DeviceBreaker (
    DeviceId INT NOT NULL,
    BreakerId INT NOT NULL,
    CONSTRAINT PK_DeviceBreaker PRIMARY KEY (DeviceId, BreakerId),
    CONSTRAINT FK_DeviceBreaker_Device FOREIGN KEY (DeviceId) REFERENCES Device(Id) ON DELETE CASCADE,
    CONSTRAINT FK_DeviceBreaker_Breaker FOREIGN KEY (BreakerId) REFERENCES Breaker(Id)
);

CREATE INDEX IX_DeviceBreaker_BreakerId ON DeviceBreaker(BreakerId);

CREATE TABLE User (
    Id INT IDENTITY(1, 1) PRIMARY KEY,
    Username NVARCHAR(127) NOT NULL UNIQUE,
    CONSTRAINT UQ_User_Username UNIQUE (Username)
);

CREATE TABLE UserSecret (
    UserId INT PRIMARY KEY,
    PasswordSalt VARBINARY(32) NOT NULL,
    PasswordHash VARBINARY(64) NOT NULL,
    FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE
);

CREATE TABLE SpaceGroupOwner (
    UserId INT PRIMARY KEY,
    SpaceGroupId INT NOT NULL,
    CONSTRAINT PK_SpaceGroupOwner PRIMARY KEY (UserId, SpaceGroupId),
    CONSTRAINT FK_SpaceGroupOwner_User FOREIGN KEY (UserId) REFERENCES [User](Id) ON DELETE CASCADE,
    CONSTRAINT FK_SpaceGroupOwner_SpaceGroup FOREIGN KEY (SpaceGroupId) REFERENCES SpaceGroup(Id) ON DELETE CASCADE
);