﻿using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Acrocuit.Migrations
{
    /// <inheritdoc />
    public partial class RevertBreakerUpstreamFkToNoAction : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Breaker_UpstreamBreaker",
                table: "Breaker");

            migrationBuilder.DropUniqueConstraint(
                name: "UQ_Breaker_Id_SpaceGroupId",
                table: "Breaker");

            migrationBuilder.DropIndex(
                name: "IX_Breaker_UpstreamBreaker",
                table: "Breaker");

            migrationBuilder.CreateIndex(
                name: "IX_Breaker_UpstreamBreaker",
                table: "Breaker",
                column: "UpstreamBreaker");

            migrationBuilder.AddForeignKey(
                name: "FK_Breaker_UpstreamBreaker",
                table: "Breaker",
                column: "UpstreamBreaker",
                principalTable: "Breaker",
                principalColumn: "Id");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Breaker_UpstreamBreaker",
                table: "Breaker");

            migrationBuilder.DropIndex(
                name: "IX_Breaker_UpstreamBreaker",
                table: "Breaker");

            migrationBuilder.AddUniqueConstraint(
                name: "UQ_Breaker_Id_SpaceGroupId",
                table: "Breaker",
                columns: new[] { "Id", "SpaceGroupId" });

            migrationBuilder.CreateIndex(
                name: "IX_Breaker_UpstreamBreaker",
                table: "Breaker",
                columns: new[] { "UpstreamBreaker", "SpaceGroupId" });

            migrationBuilder.AddForeignKey(
                name: "FK_Breaker_UpstreamBreaker",
                table: "Breaker",
                columns: new[] { "UpstreamBreaker", "SpaceGroupId" },
                principalTable: "Breaker",
                principalColumns: new[] { "Id", "SpaceGroupId" });
        }
    }
}
