using Microsoft.AspNetCore.Http;
using Npgsql;

namespace RoomBooking.SharedKernel.Database;

public static class PgErrors
{
    public const string UniqueViolation = "23505";
    public const string ForeignKeyViolation = "23503";
    public const string CheckViolation = "23514";
    public const string ExclusionViolation = "23P01";

    public static PostgresException? Find(Exception? ex)
    {
        while (ex is not null)
        {
            if (ex is PostgresException pg) return pg;
            ex = ex.InnerException;
        }
        return null;
    }
    public static bool IsKnown(Exception ex) =>
        Find(ex) is { SqlState: var s } &&
        (s is UniqueViolation or ForeignKeyViolation or CheckViolation or ExclusionViolation
         || s.StartsWith("BK", StringComparison.Ordinal));

    public static IResult ToProblem(Exception ex)
    {
        var pg = Find(ex) ?? throw new InvalidOperationException("Это не ошибка PostgreSQL.", ex);

        return pg.SqlState switch
        {
            UniqueViolation => Results.Problem(statusCode: 409, title: "Конфликт",
                detail: $"Запись уже существует или слот занят (ограничение {pg.ConstraintName})."),
            ExclusionViolation => Results.Problem(statusCode: 409, title: "Пересечение по времени",
                detail: $"Нарушено ограничение {pg.ConstraintName}."),
            CheckViolation => Results.Problem(statusCode: 400, title: "Некорректные данные",
                detail: $"Нарушено ограничение {pg.ConstraintName}."),
            ForeignKeyViolation => Results.Problem(statusCode: 404, title: "Связанная запись не найдена",
                detail: pg.ConstraintName is null ? pg.MessageText : $"Нарушена связь {pg.ConstraintName}."),
            var s when s.StartsWith("BK", StringComparison.Ordinal) => Results.Problem(statusCode: 422,
                title: "Нарушено правило бронирования", detail: pg.MessageText),
            _ => Results.Problem(statusCode: 500, title: "Ошибка базы данных")
        };
    }
}
