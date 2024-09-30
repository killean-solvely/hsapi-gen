export class HubspotClient {
  private hubspotClient: hubspot.Client;

  constructor(token?: string) {
    if (!token) {
      throw new Error("No token provided");
    }

    this.hubspotClient = new hubspot.Client({
      accessToken: token,
    });
  }

  private getObjectTypeFunction<T extends keyof ObjectTypes>(
    type: keyof ObjectTypes,
  ) {
    return async  <K extends keyof ObjectTypes[T]>(
      objectId: string,
      properties: K[],
    ): Promise<Pick<ObjectTypes[T], K>> => {
      const res = await this.hubspotClient.crm.objects.basicApi.getById(
        type,
        objectId,
        properties as string[],
      );

      const propResults: Pick<ObjectTypes[T], K> = res.properties as Pick<
        ObjectTypes[T],
        K
      >;

      return propResults;
    };
  }

  private createObjectTypeFunction<T extends keyof ObjectTypes>(
    type: keyof ObjectTypes,
  ) {
    return async <K extends keyof ObjectTypes[T]>(
      properties: Pick<ObjectTypes[T], K>
    ): Promise<Pick<ObjectTypes[T], K> & {
      /** The unique ID for this record. This value is set automatically by HubSpot. */
      hs_object_id: string
    }> => {
      const res = await this.hubspotClient.crm.objects.basicApi.create(type, {
        properties: properties as Record<string, string>,
        associations: [],
      });

      const propResults = {} as Pick<ObjectTypes[T], K>;

      for (const key of Object.keys(properties) as K[]) {
        if (res.properties[key as string] !== undefined) {
          propResults[key] = res.properties[key as string] as ObjectTypes[T][K];
        }
      }

      return { ...propResults, hs_object_id: res.id };
    };
  }

  private updateObjectTypeFunction<T extends keyof ObjectTypes>(
    type: keyof ObjectTypes,
  ) {
    return async (
      objectId: string,
      properties: Partial<ObjectTypes[T]>,
    ) => {
      await this.hubspotClient.crm.objects.basicApi.update(type, objectId, {
        properties: {
          ...(properties as Record<string, string>),
        },
      });
      return {};
    };
  }

  private getAssociationsObjectTypeFunction<T extends TypeKeys>(
    sourceType: T
  ) {
    return async (
      fromObjID: string,
      toObjType: TypeKeys,
    ): Promise<MultiAssociatedObjectWithLabel[]> => {
      const fromTypeID = TypeToObjectIDList[sourceType];
      const toTypeID = TypeToObjectIDList[toObjType];

      const result = await this.hubspotClient.crm.associations.v4.basicApi.getPage(fromTypeID, fromObjID, toTypeID);
      return result.results;
    };
  }

  private associateObjectTypeFunction<FromObjType extends keyof AssociationsConfig>(
    sourceType: FromObjType
  ) {
    return async <ToObjType extends keyof AssociationsConfig[FromObjType] & TypeKeys>(
      fromObjID: string,
      toObjID: string,
      toObjType: ToObjType,
      associationType: AssociationKeys<FromObjType, ToObjType>
    ): Promise<void> => {
      const fromTypeID = TypeToObjectIDList[sourceType];
      const toTypeID = TypeToObjectIDList[toObjType];
      const assocDetails = AssociationsConfig[sourceType][toObjType][associationType] as {
        ID: number;
        Category: AssociationSpecAssociationCategoryEnum;
      };

      if (!assocDetails) throw new Error("Invalid association type");

      await this.hubspotClient.crm.associations.v4.basicApi.create(
        fromTypeID,
        fromObjID,
        toTypeID,
        toObjID,
        [{
          associationTypeId: assocDetails.ID,
          associationCategory: assocDetails.Category,
        }]
      );
    };
  }
