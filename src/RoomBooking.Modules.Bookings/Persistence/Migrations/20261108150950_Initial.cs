using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace RoomBooking.Modules.Bookings.Persistence.Migrations
{
    public partial class Initial : Migration
    {
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.Sql(
                "CREATE EXTENSION IF NOT EXISTS btree_gist;");

            migrationBuilder.EnsureSchema(
                name: "bookings");

            migrationBuilder.CreateTable(
                name: "bookings",
                schema: "bookings",
                columns: table => new
                {
                    id = table.Column<Guid>(
                        type: "uuid",
                        nullable: false),

                    room_id = table.Column<Guid>(
                        type: "uuid",
                        nullable: false),

                    user_id = table.Column<Guid>(
                        type: "uuid",
                        nullable: false),

                    start_at = table.Column<DateTimeOffset>(
                        type: "timestamp with time zone",
                        nullable: false),

                    end_at = table.Column<DateTimeOffset>(
                        type: "timestamp with time zone",
                        nullable: false),

                    status = table.Column<string>(
                        type: "character varying(16)",
                        maxLength: 16,
                        nullable: false),

                    conference_url = table.Column<string>(
                        type: "character varying(500)",
                        maxLength: 500,
                        nullable: true),

                    created_at = table.Column<DateTimeOffset>(
                        type: "timestamp with time zone",
                        nullable: false),

                    cancelled_at = table.Column<DateTimeOffset>(
                        type: "timestamp with time zone",
                        nullable: true),

                    reminder_sent_at = table.Column<DateTimeOffset>(
                        type: "timestamp with time zone",
                        nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey(
                        "pk_bookings",
                        x => x.id);

                    table.CheckConstraint(
                        "ck_bookings_aligned",
                        "mod(extract(epoch from start_at), 1800) = 0 AND mod(extract(epoch from end_at), 1800) = 0");

                    table.CheckConstraint(
                        "ck_bookings_cancelled_at",
                        "(status = 'cancelled') = (cancelled_at IS NOT NULL)");

                    table.CheckConstraint(
                        "ck_bookings_end_after_start",
                        "end_at > start_at");

                    table.CheckConstraint(
                        "ck_bookings_max_duration",
                        "end_at - start_at <= interval '4 hours'");

                    table.CheckConstraint(
                        "ck_bookings_status",
                        "status IN ('active','cancelled')");
                });

            migrationBuilder.CreateIndex(
                name: "ix_bookings_pending_reminders",
                schema: "bookings",
                table: "bookings",
                column: "start_at",
                filter: "status = 'active' AND reminder_sent_at IS NULL");

            migrationBuilder.CreateIndex(
                name: "ix_bookings_user_start",
                schema: "bookings",
                table: "bookings",
                columns: new[] { "user_id", "start_at" });

            migrationBuilder.Sql("""
                ALTER TABLE bookings.bookings
                    ADD CONSTRAINT ex_bookings_no_overlap
                    EXCLUDE USING gist (
                        room_id WITH =,
                        tstzrange(start_at, end_at, '[)') WITH &&
                    )
                    WHERE (status = 'active');
                """);

            migrationBuilder.Sql("""
                ALTER TABLE bookings.bookings
                    ADD CONSTRAINT fk_bookings_room
                        FOREIGN KEY (room_id)
                        REFERENCES rooms.rooms (id)
                        ON DELETE RESTRICT,
                    ADD CONSTRAINT fk_bookings_user
                        FOREIGN KEY (user_id)
                        REFERENCES identity.users (id)
                        ON DELETE RESTRICT;
                """);

            migrationBuilder.Sql("""
                CREATE OR REPLACE FUNCTION bookings.fn_bookings_guard()
                RETURNS trigger
                LANGUAGE plpgsql AS $$
                DECLARE
                    v_start     timestamp;
                    v_start_min int;
                    v_end_min   int;
                    v_should_validate boolean;
                BEGIN
                    IF TG_OP = 'INSERT' THEN
                        IF NEW.status = 'active' AND NEW.start_at <= now() THEN
                            RAISE EXCEPTION 'Нельзя бронировать время в прошлом.'
                                USING ERRCODE = 'BK001';
                        END IF;

                        v_should_validate := NEW.status = 'active';
                    ELSE
                        IF OLD.status = 'cancelled'
                           AND NEW.status = 'active' THEN
                            RAISE EXCEPTION 'Отменённую бронь нельзя восстановить.'
                                USING ERRCODE = 'BK005';
                        END IF;

                        IF OLD.status = 'active'
                           AND NEW.status = 'cancelled'
                           AND OLD.start_at <= now() THEN
                            RAISE EXCEPTION 'Нельзя отменить уже начавшуюся встречу.'
                                USING ERRCODE = 'BK004';
                        END IF;

                        v_should_validate :=
                            NEW.status = 'active'
                            AND (
                                NEW.room_id IS DISTINCT FROM OLD.room_id
                                OR NEW.start_at IS DISTINCT FROM OLD.start_at
                                OR NEW.end_at IS DISTINCT FROM OLD.end_at
                            );
                    END IF;

                    IF v_should_validate THEN
                        IF NOT EXISTS (
                            SELECT 1
                            FROM rooms.rooms r
                            WHERE r.id = NEW.room_id
                              AND r.is_active
                        ) THEN
                            RAISE EXCEPTION 'Комната не найдена или неактивна.'
                                USING ERRCODE = 'BK002';
                        END IF;

                        v_start := NEW.start_at AT TIME ZONE 'UTC';

                        v_start_min :=
                            (
                                extract(hour FROM v_start) * 60
                                + extract(minute FROM v_start)
                            )::int;

                        v_end_min :=
                            v_start_min
                            + (
                                extract(epoch FROM (NEW.end_at - NEW.start_at)) / 60
                            )::int;

                        IF NOT EXISTS (
                            SELECT 1
                            FROM rooms.room_schedule_rules s
                            WHERE s.room_id = NEW.room_id
                              AND s.day_of_week = extract(dow FROM v_start)::int
                              AND s.start_minute <= v_start_min
                              AND s.end_minute >= v_end_min
                        ) THEN
                            RAISE EXCEPTION 'Бронь выходит за расписание доступности комнаты.'
                                USING ERRCODE = 'BK003';
                        END IF;
                    END IF;

                    RETURN NEW;
                END $$;
                """);

            migrationBuilder.Sql("""
                CREATE TRIGGER trg_bookings_guard
                BEFORE INSERT OR UPDATE OF status, room_id, start_at, end_at
                ON bookings.bookings
                FOR EACH ROW
                EXECUTE FUNCTION bookings.fn_bookings_guard();
                """);
        }

        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.Sql("""
                DROP TRIGGER IF EXISTS trg_bookings_guard
                ON bookings.bookings;
                """);

            migrationBuilder.Sql("""
                DROP FUNCTION IF EXISTS bookings.fn_bookings_guard();
                """);

            migrationBuilder.DropTable(
                name: "bookings",
                schema: "bookings");
        }
    }
}