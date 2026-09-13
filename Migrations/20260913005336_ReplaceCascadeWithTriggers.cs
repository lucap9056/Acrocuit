﻿using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Acrocuit.Migrations
{
    /// <inheritdoc />
    public partial class ReplaceCascadeWithTriggers : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Breaker_BreakerGroup",
                table: "Breaker");

            migrationBuilder.DropForeignKey(
                name: "FK_BreakerGroup_Space",
                table: "BreakerGroup");

            migrationBuilder.DropForeignKey(
                name: "FK_Space_SpaceGroup",
                table: "Space");

            migrationBuilder.AddForeignKey(
                name: "FK_Breaker_BreakerGroup",
                table: "Breaker",
                columns: new[] { "BreakerGroupId", "SpaceGroupId" },
                principalTable: "BreakerGroup",
                principalColumns: new[] { "Id", "SpaceGroupId" });

            migrationBuilder.AddForeignKey(
                name: "FK_BreakerGroup_Space",
                table: "BreakerGroup",
                columns: new[] { "SpaceId", "SpaceGroupId" },
                principalTable: "Space",
                principalColumns: new[] { "Id", "SpaceGroupId" });

            migrationBuilder.AddForeignKey(
                name: "FK_Space_SpaceGroup",
                table: "Space",
                column: "SpaceGroupId",
                principalTable: "SpaceGroup",
                principalColumn: "Id");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Breaker_BreakerGroup",
                table: "Breaker");

            migrationBuilder.DropForeignKey(
                name: "FK_BreakerGroup_Space",
                table: "BreakerGroup");

            migrationBuilder.DropForeignKey(
                name: "FK_Space_SpaceGroup",
                table: "Space");

            migrationBuilder.AddForeignKey(
                name: "FK_Breaker_BreakerGroup",
                table: "Breaker",
                columns: new[] { "BreakerGroupId", "SpaceGroupId" },
                principalTable: "BreakerGroup",
                principalColumns: new[] { "Id", "SpaceGroupId" },
                onDelete: ReferentialAction.Cascade);

            migrationBuilder.AddForeignKey(
                name: "FK_BreakerGroup_Space",
                table: "BreakerGroup",
                columns: new[] { "SpaceId", "SpaceGroupId" },
                principalTable: "Space",
                principalColumns: new[] { "Id", "SpaceGroupId" },
                onDelete: ReferentialAction.Cascade);

            migrationBuilder.AddForeignKey(
                name: "FK_Space_SpaceGroup",
                table: "Space",
                column: "SpaceGroupId",
                principalTable: "SpaceGroup",
                principalColumn: "Id",
                onDelete: ReferentialAction.Cascade);
        }
    }
}
