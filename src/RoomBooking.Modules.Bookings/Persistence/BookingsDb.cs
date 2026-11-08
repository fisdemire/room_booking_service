using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Bookings.Domain;

namespace RoomBooking.Modules.Bookings.Persistence;

internal sealed class BookingsDb(DbContextOptions<BookingsDb> options) : DbContext(options)
{
    public DbSet<Booking> Bookings => Set<Booking>();

    protected override void OnModelCreating(ModelBuilder b)
    {
        b.HasDefaultSchema("bookings");

        b.Entity<Booking>(e =>
        {
            e.ToTable("bookings", t =>
            {
                t.HasCheckConstraint("ck_bookings_status", "status IN ('active','cancelled')");
                t.HasCheckConstraint("ck_bookings_end_after_start", "end_at > start_at");
                t.HasCheckConstraint("ck_bookings_aligned",
                    "mod(extract(epoch from start_at), 1800) = 0 AND mod(extract(epoch from end_at), 1800) = 0");
                t.HasCheckConstraint("ck_bookings_max_duration", "end_at - start_at <= interval '4 hours'");
                t.HasCheckConstraint("ck_bookings_cancelled_at",
                    "(status = 'cancelled') = (cancelled_at IS NOT NULL)");
            });

            e.HasKey(x => x.Id);
            e.Property(x => x.Id).ValueGeneratedNever();
            e.Property(x => x.Status).HasMaxLength(16);
            e.Property(x => x.ConferenceUrl).HasMaxLength(500);

            e.HasIndex(x => new { x.UserId, x.StartAt }).HasDatabaseName("ix_bookings_user_start");
            e.HasIndex(x => x.StartAt).HasDatabaseName("ix_bookings_pending_reminders")
                .HasFilter("status = 'active' AND reminder_sent_at IS NULL");
        });
    }
}
