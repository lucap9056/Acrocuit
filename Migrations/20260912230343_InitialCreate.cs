﻿using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Acrocuit.Migrations
{
    /// <inheritdoc />
    public partial class InitialCreate : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "SpaceGroup",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(255)", maxLength: 255, nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_SpaceGroup", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "User",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Username = table.Column<string>(type: "nvarchar(127)", maxLength: 127, nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_User", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Space",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(255)", maxLength: 255, nullable: false),
                    SpaceGroupId = table.Column<int>(type: "int", nullable: false),
                    DisplayOrder = table.Column<int>(type: "int", nullable: false),
                    BackgroundImageUpdatedAt = table.Column<DateTime>(type: "datetime2", nullable: false, defaultValueSql: "GETUTCDATE()")
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Space", x => x.Id);
                    table.UniqueConstraint("UQ_Space_Id_SpaceGroupId", x => new { x.Id, x.SpaceGroupId });
                    table.ForeignKey(
                        name: "FK_Space_SpaceGroup",
                        column: x => x.SpaceGroupId,
                        principalTable: "SpaceGroup",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "SpaceGroupOwner",
                columns: table => new
                {
                    UserId = table.Column<int>(type: "int", nullable: false),
                    SpaceGroupId = table.Column<int>(type: "int", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_SpaceGroupOwner", x => new { x.UserId, x.SpaceGroupId });
                    table.ForeignKey(
                        name: "FK_SpaceGroupOwner_SpaceGroup",
                        column: x => x.SpaceGroupId,
                        principalTable: "SpaceGroup",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_SpaceGroupOwner_User",
                        column: x => x.UserId,
                        principalTable: "User",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "UserSecret",
                columns: table => new
                {
                    UserId = table.Column<int>(type: "int", nullable: false),
                    PasswordSalt = table.Column<byte[]>(type: "varbinary(32)", maxLength: 32, nullable: false),
                    PasswordHash = table.Column<byte[]>(type: "varbinary(64)", maxLength: 64, nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UserSecret", x => x.UserId);
                    table.ForeignKey(
                        name: "FK_UserSecret_User_UserId",
                        column: x => x.UserId,
                        principalTable: "User",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "BreakerGroup",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(255)", maxLength: 255, nullable: false),
                    SpaceId = table.Column<int>(type: "int", nullable: false),
                    SpaceGroupId = table.Column<int>(type: "int", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_BreakerGroup", x => x.Id);
                    table.UniqueConstraint("UQ_BreakerGroup_Id_SpaceGroupId", x => new { x.Id, x.SpaceGroupId });
                    table.ForeignKey(
                        name: "FK_BreakerGroup_Space",
                        columns: x => new { x.SpaceId, x.SpaceGroupId },
                        principalTable: "Space",
                        principalColumns: new[] { "Id", "SpaceGroupId" },
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Device",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(255)", maxLength: 255, nullable: false),
                    SpaceId = table.Column<int>(type: "int", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Device", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Device_Space",
                        column: x => x.SpaceId,
                        principalTable: "Space",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Breaker",
                columns: table => new
                {
                    Id = table.Column<int>(type: "int", nullable: false)
                        .Annotation("SqlServer:Identity", "1, 1"),
                    Name = table.Column<string>(type: "nvarchar(255)", maxLength: 255, nullable: false),
                    BreakerGroupId = table.Column<int>(type: "int", nullable: false),
                    SpaceGroupId = table.Column<int>(type: "int", nullable: false),
                    DisplayOrder = table.Column<int>(type: "int", nullable: false),
                    UpstreamBreaker = table.Column<int>(type: "int", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Breaker", x => x.Id);
                    table.UniqueConstraint("UQ_Breaker_Id_BreakerGroupId", x => new { x.Id, x.BreakerGroupId });
                    table.UniqueConstraint("UQ_Breaker_Id_SpaceGroupId", x => new { x.Id, x.SpaceGroupId });
                    table.CheckConstraint("CHK_Breaker_NotSelfUpstream", "[UpstreamBreaker] IS NULL OR [UpstreamBreaker] <> [Id]");
                    table.ForeignKey(
                        name: "FK_Breaker_BreakerGroup",
                        columns: x => new { x.BreakerGroupId, x.SpaceGroupId },
                        principalTable: "BreakerGroup",
                        principalColumns: new[] { "Id", "SpaceGroupId" },
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_Breaker_UpstreamBreaker",
                        columns: x => new { x.UpstreamBreaker, x.SpaceGroupId },
                        principalTable: "Breaker",
                        principalColumns: new[] { "Id", "SpaceGroupId" });
                });

            migrationBuilder.CreateTable(
                name: "BreakerGroupPos",
                columns: table => new
                {
                    BreakerGroupId = table.Column<int>(type: "int", nullable: false),
                    PositionX = table.Column<int>(type: "int", nullable: false),
                    PositionY = table.Column<int>(type: "int", nullable: false),
                    PositionZ = table.Column<int>(type: "int", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_BreakerGroupPos", x => x.BreakerGroupId);
                    table.ForeignKey(
                        name: "FK_BreakerGroupPos_BreakerGroup_BreakerGroupId",
                        column: x => x.BreakerGroupId,
                        principalTable: "BreakerGroup",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "DevicePos",
                columns: table => new
                {
                    DeviceId = table.Column<int>(type: "int", nullable: false),
                    PositionX = table.Column<int>(type: "int", nullable: false),
                    PositionY = table.Column<int>(type: "int", nullable: false),
                    PositionZ = table.Column<int>(type: "int", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_DevicePos", x => x.DeviceId);
                    table.ForeignKey(
                        name: "FK_DevicePos_Device_DeviceId",
                        column: x => x.DeviceId,
                        principalTable: "Device",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "DeviceBreaker",
                columns: table => new
                {
                    DeviceId = table.Column<int>(type: "int", nullable: false),
                    BreakerId = table.Column<int>(type: "int", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_DeviceBreaker", x => new { x.DeviceId, x.BreakerId });
                    table.ForeignKey(
                        name: "FK_DeviceBreaker_Breaker",
                        column: x => x.BreakerId,
                        principalTable: "Breaker",
                        principalColumn: "Id");
                    table.ForeignKey(
                        name: "FK_DeviceBreaker_Device",
                        column: x => x.DeviceId,
                        principalTable: "Device",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Breaker_BreakerGroupId",
                table: "Breaker",
                column: "BreakerGroupId");

            migrationBuilder.CreateIndex(
                name: "IX_Breaker_BreakerGroupId_SpaceGroupId",
                table: "Breaker",
                columns: new[] { "BreakerGroupId", "SpaceGroupId" });

            migrationBuilder.CreateIndex(
                name: "IX_Breaker_UpstreamBreaker",
                table: "Breaker",
                columns: new[] { "UpstreamBreaker", "SpaceGroupId" });

            migrationBuilder.CreateIndex(
                name: "UQ_BreakerGroup_Order",
                table: "Breaker",
                columns: new[] { "BreakerGroupId", "DisplayOrder" },
                unique: true);

            migrationBuilder.CreateIndex(
                name: "IX_BreakerGroup_SpaceId",
                table: "BreakerGroup",
                column: "SpaceId");

            migrationBuilder.CreateIndex(
                name: "IX_BreakerGroup_SpaceId_SpaceGroupId",
                table: "BreakerGroup",
                columns: new[] { "SpaceId", "SpaceGroupId" });

            migrationBuilder.CreateIndex(
                name: "IX_Device_SpaceId",
                table: "Device",
                column: "SpaceId");

            migrationBuilder.CreateIndex(
                name: "IX_DeviceBreaker_BreakerId",
                table: "DeviceBreaker",
                column: "BreakerId");

            migrationBuilder.CreateIndex(
                name: "IX_Space_SpaceGroupId",
                table: "Space",
                column: "SpaceGroupId");

            migrationBuilder.CreateIndex(
                name: "UQ_SpaceGroup_Order",
                table: "Space",
                columns: new[] { "SpaceGroupId", "DisplayOrder" },
                unique: true);

            migrationBuilder.CreateIndex(
                name: "IX_SpaceGroupOwner_SpaceGroupId",
                table: "SpaceGroupOwner",
                column: "SpaceGroupId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "BreakerGroupPos");

            migrationBuilder.DropTable(
                name: "DeviceBreaker");

            migrationBuilder.DropTable(
                name: "DevicePos");

            migrationBuilder.DropTable(
                name: "SpaceGroupOwner");

            migrationBuilder.DropTable(
                name: "UserSecret");

            migrationBuilder.DropTable(
                name: "Breaker");

            migrationBuilder.DropTable(
                name: "Device");

            migrationBuilder.DropTable(
                name: "User");

            migrationBuilder.DropTable(
                name: "BreakerGroup");

            migrationBuilder.DropTable(
                name: "Space");

            migrationBuilder.DropTable(
                name: "SpaceGroup");
        }
    }
}
