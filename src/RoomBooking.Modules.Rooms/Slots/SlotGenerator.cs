using RoomBooking.Modules.Rooms.Domain;

namespace RoomBooking.Modules.Rooms.Slots;

internal static class SlotGenerator
{
    public const int SlotMinutes = 30;

    public static IEnumerable<TimeSlot> Generate(
        IReadOnlyCollection<RoomScheduleRule> rules, DateOnly from, DateOnly to)
    {
        for (var date = from; date <= to; date = date.AddDays(1))
        {
            var dayStart = new DateTimeOffset(date.ToDateTime(TimeOnly.MinValue, DateTimeKind.Utc));

            foreach (var r in rules.Where(x => x.DayOfWeek == date.DayOfWeek).OrderBy(x => x.StartMinute))
                for (var m = (int)r.StartMinute; m + SlotMinutes <= r.EndMinute; m += SlotMinutes)
                    yield return new TimeSlot(dayStart.AddMinutes(m), dayStart.AddMinutes(m + SlotMinutes));
        }
    }
}