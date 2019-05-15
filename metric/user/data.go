package user

const (
	EventUserOpenApp = "EventUserOpenApp"
)

const (
	OpenAppTimeDistributionSlotCounter  = "OpenAppTimeDistributionSlotCounter"
	OpenAppCountDistributionSlotCounter = "OpenAppCountDistributionSlotCounter"
	OpenAppCPVCounter                   = "OpenAppCPVCounter"
	NewUserTimeDistributionSlotCounter  = "NewUserTimeDistributionSlotCounter"
	NewRegisteredUserCPVCounter         = "NewRegisteredUserCPVCounter"
	NewUserCPVCounter                   = "NewUserCPVCounter"
	DailyActiveCPVCounter               = "DailyActiveCPVCounter"

	NewUserRetentionSlotCounter              = "NewUserRetentionSlotCounter"
	ChannelNewUserRetentionSlotCounterPrefix = "channelNewUserRetentionSlotCounter_"

	ActiveUserRetentionSlotCounter              = "ActiveUserRetentionSlotCounter"
	ChannelActiveUserRetentionSlotCounterPrefix = "channelActiveUserRetentionSlotCounter_"
	ActiveUserTimeDistributionSlotCounter       = "ActiveUserTimeDistributionSlotCounter"

	DailyActiveNewUserPercentSimpleCounter = "DailyActiveNewUserPercentSimpleCounter"
	DailyActiveUserAffinitySlotCounter     = "DailyActiveUserAffinitySlotCounter"
	DailyActiveUserFreshnessSlotCounter    = "DailyActiveUserFreshnessSlotCounter"
)

const (
	DailyScheduleEvent = "UserDailyScheduleEvent"
)

var (
	retentionDays                 = [...]int{1, 2, 3, 4, 5, 6, 7, 15, 30}
	OpenAppCountDistributionSlots = []string{"1-2", "3-4", "5-6", "7-8", "9-10", "11-20", "21-30", "31-49", "50+"}
)
