﻿using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Acrocuit.Migrations
{
    /// <inheritdoc />
    public partial class AddUserUsernameUniqueIndex : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateIndex(
                name: "UQ_User_Username",
                table: "User",
                column: "Username",
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropIndex(
                name: "UQ_User_Username",
                table: "User");
        }
    }
}
