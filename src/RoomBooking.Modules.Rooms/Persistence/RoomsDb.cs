using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Rooms.Domain;

namespace RoomBooking.Modules.Rooms.Persistence;

internal sealed class RoomsDb(DbContextOptions<RoomsDb> options) : DbContext(options)
{
    public DbSet<Room> Rooms => Set<Room>();
    public DbSet<RoomScheduleRule> ScheduleRules => Set<RoomScheduleRule>();

    protected override void OnModelCreating(ModelBuilder b)
    {
        b.HasDefaultSchema("rooms");

        b.Entity<Room>(e =>
        {
            e.ToTable("rooms", t =>
            {
                t.HasCheckConstraint("ck_rooms_name_not_blank", "length(btrim(name)) > 0");
                t.HasCheckConstraint("ck_rooms_capacity", "capacity BETWEEN 1 AND 500");
            });
            e.HasKey(x => x.Id);
            e.Property(x => x.Id).ValueGeneratedNever();
            e.Property(x => x.Name).HasMaxLength(120);
            e.Property(x => x.CreatedAt).HasDefaultValueSql("now()");
            e.HasIndex(x => x.Name).IsUnique().HasDatabaseName("ux_rooms_name");
        });

        b.Entity<RoomScheduleRule>(e =>
        {
            e.ToTable("room_schedule_rules", t =>
            {
                t.HasCheckConstraint("ck_rule_day", "day_of_week BETWEEN 0 AND 6");
                t.HasCheckConstraint("ck_rule_start", "start_minute BETWEEN 0 AND 1410 AND start_minute % 30 = 0");
                t.HasCheckConstraint("ck_rule_end", "end_minute BETWEEN 30 AND 1440 AND end_minute % 30 = 0");
                t.HasCheckConstraint("ck_rule_order", "end_minute > start_minute");
            });
            e.HasKey(x => x.Id);
            e.Property(x => x.Id).ValueGeneratedNever();
            e.Property(x => x.DayOfWeek).HasConversion<short>();
            e.HasOne<Room>().WithMany().HasForeignKey(x => x.RoomId).OnDelete(DeleteBehavior.Cascade);
            e.HasIndex(x => x.RoomId);
        });
    }
}
