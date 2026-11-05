using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace RoomBooking.Modules.Rooms.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class Initial : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.Sql("CREATE EXTENSION IF NOT EXISTS btree_gist;");

            migrationBuilder.EnsureSchema(
                name: "rooms");

            migrationBuilder.CreateTable(
                name: "rooms",
                schema: "rooms",
                columns: table => new
                {
                    id = table.Column<Guid>(type: "uuid", nullable: false),
                    name = table.Column<string>(type: "character varying(120)", maxLength: 120, nullable: false),
                    capacity = table.Column<int>(type: "integer", nullable: false),
                    is_active = table.Column<bool>(type: "boolean", nullable: false),
                    created_at = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: false, defaultValueSql: "now()")
                },
                constraints: table =>
                {
                    table.PrimaryKey("pk_rooms", x => x.id);
                    table.CheckConstraint("ck_rooms_capacity", "capacity BETWEEN 1 AND 500");
                    table.CheckConstraint("ck_rooms_name_not_blank", "length(btrim(name)) > 0");
                });

            migrationBuilder.CreateTable(
                name: "room_schedule_rules",
                schema: "rooms",
                columns: table => new
                {
                    id = table.Column<Guid>(type: "uuid", nullable: false),
                    room_id = table.Column<Guid>(type: "uuid", nullable: false),
                    day_of_week = table.Column<short>(type: "smallint", nullable: false),
                    start_minute = table.Column<short>(type: "smallint", nullable: false),
                    end_minute = table.Column<short>(type: "smallint", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("pk_room_schedule_rules", x => x.id);
                    table.CheckConstraint("ck_rule_day", "day_of_week BETWEEN 0 AND 6");
                    table.CheckConstraint("ck_rule_end", "end_minute BETWEEN 30 AND 1440 AND end_minute % 30 = 0");
                    table.CheckConstraint("ck_rule_order", "end_minute > start_minute");
                    table.CheckConstraint("ck_rule_start", "start_minute BETWEEN 0 AND 1410 AND start_minute % 30 = 0");
                    table.ForeignKey(
                        name: "fk_room_schedule_rules_rooms_room_id",
                        column: x => x.room_id,
                        principalSchema: "rooms",
                        principalTable: "rooms",
                        principalColumn: "id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateIndex(
                name: "ix_room_schedule_rules_room_id",
                schema: "rooms",
                table: "room_schedule_rules",
                column: "room_id");

            migrationBuilder.CreateIndex(
                name: "ux_rooms_name",
                schema: "rooms",
                table: "rooms",
                column: "name",
                unique: true);

            migrationBuilder.Sql("""
            ALTER TABLE rooms.room_schedule_rules
                ADD CONSTRAINT ex_schedule_no_overlap
                EXCLUDE USING gist (
                    room_id WITH =,
                    day_of_week WITH =,
                    int4range(start_minute::int, end_minute::int) WITH &&
                );
            """);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "room_schedule_rules",
                schema: "rooms");

            migrationBuilder.DropTable(
                name: "rooms",
                schema: "rooms");
        }
    }
}
