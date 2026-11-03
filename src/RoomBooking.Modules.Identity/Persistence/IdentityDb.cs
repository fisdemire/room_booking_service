using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Identity.Domain;

namespace RoomBooking.Modules.Identity.Persistence;

internal sealed class IdentityDb(DbContextOptions<IdentityDb> options) : DbContext(options)
{
    public DbSet<User> Users => Set<User>();

    protected override void OnModelCreating(ModelBuilder b)
    {
        b.HasDefaultSchema("identity");

        b.Entity<User>(e =>
        {
            e.ToTable("users", t =>
            {
                t.HasCheckConstraint("ck_users_email_lowercase", "email = lower(email)");
                t.HasCheckConstraint("ck_users_email_not_blank", "length(btrim(email)) > 3");
                t.HasCheckConstraint("ck_users_role", "role IN ('Admin','User')");
            });

            e.HasKey(x => x.Id);
            e.Property(x => x.Id).ValueGeneratedNever();
            e.Property(x => x.Email).HasMaxLength(256);
            e.Property(x => x.PasswordHash).HasMaxLength(512);
            e.Property(x => x.Role).HasMaxLength(16);
            e.Property(x => x.CreatedAt).HasDefaultValueSql("now()");

            e.HasIndex(x => x.Email).IsUnique().HasDatabaseName("ux_users_email");
        });
    }
}
