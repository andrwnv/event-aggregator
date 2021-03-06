package usecases

import (
	"github.com/andrwnv/event-aggregator/core/dto"
	"github.com/andrwnv/event-aggregator/core/repo"
	"github.com/andrwnv/event-aggregator/core/services"
	"github.com/google/uuid"
)

type EventUsecase struct {
	eventRepo   *repo.EventRepo
	userUsecase *UserUsecase
	regionRepo  *repo.RegionRepo
	esService   *services.EsService
}

func NewEventUsecase(
	eventRepo *repo.EventRepo,
	userUsecase *UserUsecase,
	regionRepo *repo.RegionRepo,
	esService *services.EsService) *EventUsecase {
	return &EventUsecase{
		eventRepo:   eventRepo,
		userUsecase: userUsecase,
		regionRepo:  regionRepo,
		esService:   esService,
	}
}

func (u *EventUsecase) Get(id uuid.UUID) Result {
	event, err := u.eventRepo.Get(id)
	eventPhotos, _ := u.eventRepo.GetImages(id)
	return Result{repo.EventToEvent(event, eventPhotos), err}
}

func (u *EventUsecase) GetFullEvent(id uuid.UUID) (repo.Event, error) {
	return u.eventRepo.Get(id)
}

func (u *EventUsecase) GetFullEventComment(id uuid.UUID) (repo.EventComment, error) {
	return u.eventRepo.GetCommentByID(id)
}

func (u *EventUsecase) GetEvents(page int, count int) Result {
	places, err := u.eventRepo.GetEvents(page, count)
	if err != nil {
		return Result{nil, MakeUsecaseError("Places not found.")}
	}

	var result []dto.EventDto
	for _, value := range places {
		eventPhotos, _ := u.eventRepo.GetImages(value.ID)
		result = append(result, repo.EventToEvent(value, eventPhotos))
	}

	return Result{result, nil}
}

func (u *EventUsecase) Create(createDto dto.CreateEvent, userInfo dto.BaseUserInfo) Result {
	user, err := u.userUsecase.GetFull(userInfo)
	if err != nil {
		return Result{nil, MakeUsecaseError("Cant find user for create event.")}
	}

	region, err := u.regionRepo.GetByRegionID(createDto.RegionID)
	if err != nil {
		return Result{nil, MakeUsecaseError("Cant find selected country.")}
	}

	// TODO: check begin, end datetime correctness for upd & create

	event, err := u.eventRepo.Create(createDto, user, region)
	err = u.esService.Create(dto.CreateAggregatorRecordDto{
		ID:           event.ID,
		LocationName: createDto.Title,
		Location: dto.LocationDto{
			Lat: createDto.Latitude,
			Lon: createDto.Longitude,
		},
		LocationType: "event",
	})

	return Result{repo.EventToEvent(event, []string{}), err}
}

func (u *EventUsecase) Update(id uuid.UUID, updateDto dto.UpdateEvent, userInfo dto.BaseUserInfo) Result {
	event, err := u.eventRepo.Get(id)
	if err != nil {
		return Result{nil, err}
	}

	if userInfo.ID != event.CreatedBy.ID.String() {
		return Result{nil, MakeUsecaseError("Isn't your event!")}
	}

	event.Region, err = u.regionRepo.GetByRegionID(updateDto.RegionID)
	if err != nil {
		return Result{nil, MakeUsecaseError("Cant find selected country.")}
	}

	// TODO: check begin, end datetime correctness for upd & create

	err = u.eventRepo.Update(event.ID, updateDto, event.Region)
	err = u.esService.Update(event.ID, dto.UpdateAggregatorRecordDto{
		LocationName: updateDto.Title,
		Location: dto.LocationDto{
			Lat: updateDto.Latitude,
			Lon: updateDto.Longitude,
		},
		LocationType: "event",
	})

	return Result{err != nil, err}
}

func (u *EventUsecase) Delete(id uuid.UUID, userInfo dto.BaseUserInfo) Result {
	event, err := u.eventRepo.Get(id)
	if err != nil {
		return Result{false, err}
	}

	if userInfo.ID != event.CreatedBy.ID.String() {
		return Result{false, MakeUsecaseError("Isn't your event!")}
	}

	err = u.eventRepo.Delete(id)
	err = u.esService.Delete(id)

	return Result{err != nil, err}
}

// ----- EventUsecase: Images -----

func (u *EventUsecase) UpdateEventImages(id uuid.UUID, userInfo dto.BaseUserInfo,
	filesToCreate []string, filesToDelete []string) Result {

	event, err := u.eventRepo.Get(id)
	if err != nil {
		return Result{false, err}
	}

	if userInfo.ID != event.CreatedBy.ID.String() {
		return Result{false, MakeUsecaseError("Isn't your event!")}
	}

	for _, url := range filesToCreate {
		err := u.eventRepo.CreateImages(event.ID, url)
		if err != nil {
			return Result{false, err}
		}
	}

	for _, url := range filesToDelete {
		err := u.eventRepo.DeleteImages(url)
		// TODO: delete photos from dir.
		if err != nil {
			return Result{false, err}
		}
	}

	return Result{true, nil}
}

// ----- EventUsecase: Comments -----

func (u *EventUsecase) CreateComment(createDto dto.CreateEventCommentDto, userInfo dto.BaseUserInfo) Result {
	user, err := u.userUsecase.GetFull(userInfo)
	if err != nil {
		return Result{nil, err}
	}
	event, err := u.eventRepo.Get(uuid.MustParse(createDto.LinkedEventID))
	if err != nil {
		return Result{false, err}
	}

	comment, err := u.eventRepo.CreateComment(createDto, user, event)
	if err != nil {
		return Result{nil, MakeUsecaseError("Failed to create comment.")}
	}

	return Result{repo.CommentToComment(comment), nil}
}

func (u *EventUsecase) GetComments(eventId uuid.UUID, page int, count int) Result {
	comments, err := u.eventRepo.GetComments(eventId, page, count)
	if err != nil {
		return Result{nil, MakeUsecaseError("Failed to create comment.")}
	}

	var result []dto.EventCommentDto
	for _, value := range comments {
		result = append(result, repo.CommentToComment(value))
	}

	total, err := u.eventRepo.GetTotalCommentsCount()
	if err != nil {
		return Result{nil, MakeUsecaseError("Cant extract total count of event comment.")}
	}

	return Result{dto.EventCommentListDto{
		Page:      int64(page),
		ListSize:  int64(count),
		TotalSize: total,
		List:      result,
	}, nil}
}

func (u *EventUsecase) DeleteComment(commentId uuid.UUID, userInfo dto.BaseUserInfo) Result {
	comment, err := u.eventRepo.GetCommentByID(commentId)
	if err != nil {
		return Result{nil, err}
	}
	if userInfo.ID != comment.CreatedBy.ID.String() {
		return Result{nil, MakeUsecaseError("Isn't your comment!")}
	}

	err = u.eventRepo.DeleteComments(commentId)
	if err != nil {
		return Result{false, MakeUsecaseError("Cant delete comment(s).")}
	}
	return Result{true, nil}
}

func (u *EventUsecase) UpdateComment(id uuid.UUID, updateDto dto.UpdateEventCommentDto, userInfo dto.BaseUserInfo) Result {
	comment, err := u.eventRepo.GetCommentByID(id)
	if err != nil {
		return Result{nil, err}
	}
	if userInfo.ID != comment.CreatedBy.ID.String() {
		return Result{nil, MakeUsecaseError("Isn't your comment!")}
	}

	err = u.eventRepo.UpdateComment(id, updateDto)
	if err != nil {
		return Result{false, MakeUsecaseError("Cant update comment(s).")}
	}
	return Result{true, nil}
}
