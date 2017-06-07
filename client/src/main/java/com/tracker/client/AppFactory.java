package com.tracker.client;

import java.util.HashMap;
import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.activities.BaseActivity;
import com.tracker.client.activities.CommonEditActivity;
import com.tracker.client.activities.CommonListActivity;
import com.tracker.client.activities.CommonSearchComponent;
import com.tracker.client.activities.settings.SettingsActivity;
import com.tracker.client.activities.UserChangeEvent;
import com.tracker.client.activities.meals.MealSearchComponent;
import com.tracker.client.activities.meals.MealStatsActivity;
import com.tracker.client.activities.signin.ConfirmEmailActivity;
import com.tracker.client.activities.signin.SignInActivity;
import com.tracker.client.jso.User;
import com.tracker.client.place.AdminPlaceHistoryMapper;
import com.tracker.client.place.BasePlace;
import com.tracker.client.place.ViewId;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.rpc.ContentRpcServiceImpl;
import com.google.gwt.activity.shared.Activity;
import com.google.gwt.activity.shared.ActivityManager;
import com.google.gwt.activity.shared.ActivityMapper;
import com.google.gwt.core.client.GWT;
import com.google.gwt.place.shared.Place;
import com.google.gwt.place.shared.PlaceController;
import com.google.gwt.place.shared.PlaceHistoryHandler;
import com.google.gwt.storage.client.Storage;
import com.google.web.bindery.event.shared.EventBus;
import com.google.web.bindery.event.shared.SimpleEventBus;

import elemental.client.Browser;
import elemental.dom.Document;
import elemental.dom.Element;

public class AppFactory {
  private static final Logger log = Logger.getLogger(AppFactory.class.getName());
  private ContentRpcService contentRpcService;
  private final EventBus eventBus = new SimpleEventBus();
  private final Storage navStorage = Storage.getSessionStorageIfSupported();

  private final AdminPlaceHistoryMapper historyMapper = GWT.create(AdminPlaceHistoryMapper.class);
  private final PlaceController placeController = new PlaceController(eventBus);
  private final PlaceHistoryHandler placeHistoryHandler = new PlaceHistoryHandler(historyMapper);
  private final ActivityManager activityManager;
  private final Map<ViewId, BaseActivity> activityMap = new HashMap<>();
  private final BasePlace defaultPlace = BasePlace.newBuilder().viewId(new ViewId("signin")).build();
  private final BasePlace nextPlace = BasePlace.newBuilder().viewId(new ViewId("meals")).build();
  private final BasePlace confirmEmailPlace = BasePlace.newBuilder().viewId(new ViewId("confirm_email")).build();
  private User user;

  public AppFactory() {
    activityManager = new ActivityManager(new ActivityMapper() {
      @Override
      public Activity getActivity(Place place) {
        return innerActivityMapper((BasePlace) place);
      }
    }, eventBus);
    historyMapper.setFactory(this);
  }

  public PlaceController getPlaceController() {
    return placeController;
  }

  public PlaceHistoryHandler getPlaceHistoryHandler() {
    return placeHistoryHandler;
  }

  public ActivityManager getActivityManager() {
    return activityManager;
  }

  public BasePlace getDefaultPlace() {
    return defaultPlace;
  }

  public BasePlace getNextPlace() {
    return nextPlace;
  }

  public BasePlace getConfirmEmailPlace() {
    return confirmEmailPlace;
  }

  public BasePlace.Tokenizer getBasePlaceTokenizer() {
    return new BasePlace.Tokenizer(this);
  }

  private Activity innerActivityMapper(BasePlace place) {
    ViewId viewId = place.getViewId();
    log.info("viewId = " + viewId);
    BaseActivity activity = activityMap.get(viewId);
    if (activity == null) {
      String activityId = viewId.getActivityId();
      if ("signin".equals(activityId)) {
        activity = new SignInActivity(this);
      } else if ("users".equals(activityId)) {
        activity = new CommonListActivity(this)
          .setApiPrefix("/v1/users/")
          .setEditViewId("user_edit")
          .setSearchComponent(new CommonSearchComponent());
      } else if ("meals".equals(activityId)) {
        activity = new CommonListActivity(this)
          .setApiPrefix("/v1/meals/")
          .setEditViewId("meal_edit")
          .setSearchComponent(new MealSearchComponent());
      } else if ("meal_edit".equals(activityId)) {
        activity = new CommonEditActivity(this)
          .setApiPrefix("/v1/meals/")
          .setApiForm("/v1/meals-form/");
      } else if ("meals_stats".equals(activityId)) {
        activity = new MealStatsActivity(this);
      } else if ("user_edit".equals(activityId)) {
        activity = new CommonEditActivity(this)
          .setApiPrefix("/v1/users/")
          .setApiForm("/v1/users-form/");
      } else if ("confirm_email".equals(activityId)) {
        activity = new ConfirmEmailActivity(this);
      } else if ("settings".equals(activityId)) {
        activity = new SettingsActivity(this);
      } else {
        throw new AssertionError("activity not found " + activityId);
      }
      activityMap.put(viewId, activity);
    }
    log.info("activity = " + activity);
    activity.updateForPlace(place);
    return activity;
  }

  public ContentRpcService getRpcService() {
    if (contentRpcService == null) {
      contentRpcService = new ContentRpcServiceImpl(this, new ContentRpcService.StatusObserver() {
        @Override
        public void onServerCameBack() {
        }

        @Override
        public void onServerWentAway() {
        }

        @Override
        public void onTaskFinished() {
          log.info("onTaskFinished ");
          cancelProgress();
        }

        @Override
        public void onTaskStarted(String description) {
          log.info("onTaskStarted " + description);
          showProgress();
        }

        private void cancelProgress() {
          Document doc = Browser.getDocument();
          Element progressEl = doc.getElementById("progress");
          progressEl.setClassName("start");
        }

        private void showProgress() {
          Document doc = Browser.getDocument();
          Element progressEl = doc.getElementById("progress");
          progressEl.setClassName("finish");
        }

      });
    }
    return contentRpcService;
  }

  public EventBus getEventBus() {
    return eventBus;
  }

  public Storage getNavStorage() {
    return navStorage;
  }

  public User getCurrentUser() {
    return user;
  }

  public void setCurrentUser(User user) {
    this.user = user;
    UserChangeEvent.fire(eventBus, user, "");
  }
}
